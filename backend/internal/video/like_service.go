package video

import (
	"context"
	"errors"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type LikeService struct {
	repo      *LikeRepository
	VideoRepo *VideoRepository
	cache     *rediscache.Client
	likeMQ    *rabbitmq.LikeMQ
	popularMQ *rabbitmq.PopularityMQ
}

// isDupKey 判断是否为 MySQL 唯一键冲突错误
func isDupKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// NewLikeService 创建 LikeService 实例
func NewLikeService(repo *LikeRepository, videoRepo *VideoRepository, cache *rediscache.Client, likeMQ *rabbitmq.LikeMQ, popularMQ *rabbitmq.PopularityMQ) *LikeService {
	return &LikeService{repo: repo, VideoRepo: videoRepo, cache: cache, likeMQ: likeMQ, popularMQ: popularMQ}
}

// Like 点赞（校验 + 消息队列/事务写入 + 热度更新）
func (ls *LikeService) Like(ctx context.Context, like *Like) error {
	if like == nil {
		return errors.New("like is null")
	}
	if like.VideoID == 0 || like.AccountID == 0 {
		return errors.New("video_id and account_id are required")
	}
	if ls.VideoRepo != nil {
		ok, err := ls.VideoRepo.IsExist(ctx, like.VideoID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("video not found")
		}
	}
	ok, err := ls.repo.IsLike(ctx, like.VideoID, like.AccountID)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("user has liked this video")
	}

	like.CreatedAt = time.Now()
	mysqlEnqueued := false
	redisEnqueued := false
	if ls.likeMQ != nil {
		if err := ls.likeMQ.Like(ctx, like.AccountID, like.VideoID); err == nil {
			mysqlEnqueued = true
		}
	}
	if ls.popularMQ != nil {
		if err := ls.popularMQ.Update(ctx, like.VideoID, 1); err == nil {
			redisEnqueued = true
		}
	}
	if mysqlEnqueued == true && redisEnqueued == true {
		return nil
	}
	if !mysqlEnqueued {
		err := ls.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Select("id").First(&Video{}, like.VideoID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("video not found")
				}
				return err
			}
			if err := tx.Create(like).Error; err != nil {
				if isDupKey(err) {
					return errors.New("user has liked this video")
				}
				return err
			}
			if err := tx.Model(&Video{}).Where("id = ?", like.VideoID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
				return err
			}
			return tx.Model(&Video{}).Where("id = ?", like.VideoID).UpdateColumn("popularity", gorm.Expr("popularity + 1")).Error
		})
		if err != nil {
			return err
		}
	}
	if !redisEnqueued {
		UpdatePopularityCache(ctx, ls.cache, like.VideoID, 1)
	}
	return nil

}

// Unlike 取消点赞（校验 + 消息队列/事务删除 + 热度更新）
func (ls *LikeService) Unlike(ctx context.Context, like *Like) error {
	if ls.VideoRepo != nil {
		ok, err := ls.VideoRepo.IsExist(ctx, like.VideoID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("video not found")
		}
	}
	ok, err := ls.repo.IsLike(ctx, like.VideoID, like.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user has not liked this video")
	}

	mysqlEnqueued := false
	redisEnqueued := false
	if ls.likeMQ != nil {
		if err := ls.likeMQ.UnLike(ctx, like.AccountID, like.VideoID); err == nil {
			mysqlEnqueued = true
		}
	}
	if ls.popularMQ != nil {
		if err := ls.popularMQ.Update(ctx, like.VideoID, -1); err == nil {
			redisEnqueued = true
		}
	}
	if mysqlEnqueued == true && redisEnqueued == true {
		return nil
	}
	if !mysqlEnqueued {
		err := ls.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			del := tx.Where("video_id = ? AND account_id = ?", like.VideoID, like.AccountID).Delete(&Like{})
			if del.Error != nil {
				return del.Error
			}
			if del.RowsAffected == 0 {
				return errors.New("user has not liked this video")
			}
			if err := tx.Model(&Video{}).Where("id = ?", like.VideoID).
				UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)")).Error; err != nil {
				return err
			}
			return tx.Model(&Video{}).Where("id = ?", like.VideoID).UpdateColumn("popularity", gorm.Expr("GREATEST(popularity - 1 ,0)")).Error
		})
		if err != nil {
			return err
		}

	}
	if !redisEnqueued {
		UpdatePopularityCache(ctx, ls.cache, like.VideoID, -1)
	}
	return nil
}

// Isliked 查询是否已点赞
func (ls *LikeService) Isliked(ctx context.Context, videoID, accountID uint) (bool, error) {
	return ls.repo.IsLike(ctx, videoID, accountID)
}

// ListLikedVideos 查询用户点赞的视频列表
func (ls *LikeService) ListLikedVideos(ctx context.Context, accountID uint) ([]Video, error) {
	return ls.repo.ListLikedVideos(ctx, accountID)
}
