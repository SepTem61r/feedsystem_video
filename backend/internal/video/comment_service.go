package video

import (
	"context"
	"errors"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"strings"

	"gorm.io/gorm"
)

type CommentService struct {
	commentRepository *CommentRepository
	cache             *rediscache.Client
	popularityMQ      *rabbitmq.PopularityMQ
	videoRepository   *VideoRepository
	commentMQ         *rabbitmq.CommentMQ
}

func NewCommentService(commentRepository *CommentRepository, cache *rediscache.Client, popularMQ *rabbitmq.PopularityMQ, videoRepository *VideoRepository, commentMQ *rabbitmq.CommentMQ) *CommentService {
	return &CommentService{
		commentRepository: commentRepository,
		cache:             cache,
		popularityMQ:      popularMQ,
		videoRepository:   videoRepository,
		commentMQ:         commentMQ,
	}
}
func (cs *CommentService) Publish(ctx context.Context, comment *Comment) error {
	if comment == nil {
		return errors.New("comment is require")
	}
	comment.Username = strings.TrimSpace(comment.Username)
	comment.Content = strings.TrimSpace(comment.Content)
	if comment.VideoID == 0 || comment.AuthorID == 0 {
		return errors.New("video_id or author_id is require")
	}
	if comment.Content == "" {
		return errors.New("content is required")
	}
	isexist, err := cs.videoRepository.IsExist(ctx, comment.VideoID)
	if err != nil {
		return err
	}
	if !isexist {
		return errors.New("video is not found")
	}

	mysqlEnqueued := false
	redisEnqueued := false
	if cs.commentMQ != nil {
		if err := cs.commentMQ.Publish(ctx, comment.Username, comment.VideoID, comment.AuthorID, comment.Content); err != nil {
			return err
		}
		mysqlEnqueued = true
	}
	if cs.popularityMQ != nil {
		if err := cs.popularityMQ.Update(ctx, comment.VideoID, 1); err != nil {
			return err
		}
		redisEnqueued = true
	}
	if mysqlEnqueued == true && redisEnqueued == true {
		return nil
	}

	if !mysqlEnqueued {
		if err := cs.commentRepository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Select("id").First(&Video{}, comment.VideoID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("video not found")
				}
				return err
			}
			if err := tx.Create(comment).Error; err != nil {
				return err
			}
			return tx.Model(&Video{}).Where("id = ?", comment.VideoID).UpdateColumn("popularity", gorm.Expr("popularity + 1")).Error

		}); err != nil {
			return err
		}
	}
	if !redisEnqueued {
		UpdatePopularityCache(ctx, cs.cache, comment.VideoID, 1)
	}
	return nil
}

func (cs *CommentService) Delete(ctx context.Context, commentID, accountID uint) error {
	comment, err := cs.commentRepository.GetById(ctx, commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return errors.New("comment not found")
	}
	if comment.AuthorID != commentID {
		return errors.New("permission denied")
	}
	if cs.commentMQ != nil {
		if err := cs.commentMQ.Delete(ctx, commentID); err == nil {
			return nil
		}
	}
	return cs.commentRepository.DeleteComment(ctx, comment)
}

func (cs *CommentService) GetAll(ctx context.Context, videoID uint) ([]Comment, error) {
	if videoID <= 0 {
		return nil, errors.New("video_id is require")
	}
	exists, err := cs.videoRepository.IsExist(ctx, videoID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("video not found")
	}
	return cs.commentRepository.GetAllComments(ctx, videoID)

}
