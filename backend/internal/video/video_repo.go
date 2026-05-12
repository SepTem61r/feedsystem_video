package video

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

// NewVideoRepository 创建 VideoRepository 实例
func NewVideoRepository(db *gorm.DB) *VideoRepository { return &VideoRepository{db: db} }

// CreateVideo 插入一条视频记录
func (vr *VideoRepository) CreateVideo(ctx context.Context, video *Video) error {
	if err := vr.db.WithContext(ctx).Create(video).Error; err != nil {
		return err
	}
	return nil
}

// CreateMsg 插入一条 outbox 消息记录
func (vr *VideoRepository) CreateMsg(ctx context.Context, msg *OutboxMsg) error {
	if err := vr.db.WithContext(ctx).Create(msg).Error; err != nil {
		return err
	}
	return nil
}

// DelVideo 按主键删除视频
func (vr *VideoRepository) DelVideo(ctx context.Context, id uint) error {
	if err := vr.db.WithContext(ctx).Delete(&Video{}, id).Error; err != nil {
		return err
	}
	return nil
}

// ListByAuthorID 按作者ID查询视频列表，按创建时间倒序
func (vr *VideoRepository) ListByAuthorID(ctx context.Context, authorID uint) ([]Video, error) {
	var videos []Video
	if err := vr.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Order("createtime desc").
		Offset(0).
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// GetByID 按主键查询单个视频
func (vr *VideoRepository) GetByID(ctx context.Context, id uint) (*Video, error) {
	var video Video
	if err := vr.db.WithContext(ctx).First(&video, id).Error; err != nil {
		return (*Video)(nil), err
	}
	return &video, nil
}

// UpdateLikeCount 更新视频的点赞总数
func (vr *VideoRepository) UpdateLikeCount(ctx context.Context, id uint, likescount int64) error {
	if err := vr.db.WithContext(ctx).Model(&Video{}).
		Where("id = ?").
		Update("likes_count", likescount).Error; err != nil {
		return err
	}
	return nil
}

// IsExist 判断视频是否存在
func (vr *VideoRepository) IsExist(ctx context.Context, id uint) (bool, error) {
	var video Video
	if err := vr.db.WithContext(ctx).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdatePopularity 更新视频热度值（覆盖写入）
func (vr *VideoRepository) UpdatePopularity(ctx context.Context, id uint, change int64) error {
	if err := vr.db.WithContext(ctx).Where("id = ?", id).
		Update("popularity", gorm.Expr("popularity + ?", change)).Error; err != nil {
		return err
	}
	return nil
}

// ChangeLikesCount 原子增减点赞数（不小于0）
func (vr *VideoRepository) ChangeLikesCount(ctx context.Context, id uint, change int64) error {
	if err := vr.db.WithContext(ctx).Model(&Video{}).Where("id = ?", id).UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count + ?, 0)", change)).Error; err != nil {
		return err
	}
	return nil
}

// ChangePopularity 原子增减热度值（不小于0）
func (vr *VideoRepository) ChangePopularity(ctx context.Context, id uint, change int64) error {
	if err := vr.db.WithContext(ctx).Model(&Video{}).Where("id = ?", id).UpdateColumn("popularity", gorm.Expr("GREATEST(popularity + ?, 0)", change)).Error; err != nil {
		return err
	}
	return nil
}
