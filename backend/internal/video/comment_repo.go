package video

import (
	"context"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}
func (cr *CommentRepository) CreateComment(ctx context.Context, comment *Comment) error {
	return cr.db.WithContext(ctx).Create(comment).Error
}

func (cr *CommentRepository) DeleteComment(ctx context.Context, comment *Comment) error {
	return cr.db.WithContext(ctx).Delete(comment).Error
}

func (cr *CommentRepository) GetAllComments(ctx context.Context, videoID uint) ([]Comment, error) {
	var comments []Comment
	err := cr.db.WithContext(ctx).Where("video_id = ?", videoID).Find(&comments).Error
	return comments, err
}

func (cr *CommentRepository) IsExist(ctx context.Context, id uint) (bool, error) {
	var comment Comment
	if err := cr.db.WithContext(ctx).Where("id = ?", id).First(&comment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (cr *CommentRepository) GetById(ctx context.Context, id uint) (*Comment, error) {
	var comment Comment
	if err := cr.db.WithContext(ctx).First(&comment, id).Error; err != nil {
		return &comment, err
	}
	return &comment, nil
}
