package video

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

// 防止重复点赞
func (lr *LikeRepository) LikeIgnoreDuplicate(ctx context.Context, like *Like) (created bool, err error) {
	if like == nil || like.VideoID == 0 || like.AccountID == 0 {
		return false, nil
	}
	err = lr.db.WithContext(ctx).Create(like).Error
	if err == nil {
		return true, nil
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return false, nil
	}
	return true, nil
}

// 精准取消点赞
func (lr *LikeRepository) DeleteByVideoAndAccount(ctx context.Context, videoID, accountID uint) (delet bool, err error) {
	if videoID == 0 || accountID == 0 {
		return false, nil
	}
	res := lr.db.WithContext(ctx).Where("video_id = ? AND account_id = ?", videoID, accountID).Delete(&Like{})
	return res.RowsAffected > 0, err
}

// 单视频点赞状态查询
func (lr *LikeRepository) IsLike(ctx context.Context, videoID, accountID uint) (bool, error) {
	var count int64
	err := lr.db.WithContext(ctx).Model(&Like{}).Where("video_id = ? AND account_id = ?", videoID, accountID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 批量视频点赞状态查询
func (lr *LikeRepository) BatchGetLiked(ctx context.Context, videoIDs []uint, accountID uint) (map[uint]bool, error) {
	likemap := make(map[uint]bool)
	if len(videoIDs) == 0 || accountID == 0 {
		return likemap, nil
	}
	var likes []Like
	err := lr.db.WithContext(ctx).Model(&Like{}).Where("video_id IN ? AND account_id = ?", videoIDs, accountID).Find(&likes).Error
	if err != nil {
		return nil, err
	}
	for _, like := range likes {
		likemap[like.VideoID] = true
	}
	return likemap, err
}

// 查询用户点赞的视频列表
func (lr *LikeRepository) ListLikedVideos(ctx context.Context, accountID uint) ([]Video, error) {
	var videos []Video
	if accountID == 0 {
		return videos, nil
	}
	err := lr.db.WithContext(ctx).Model(&Video{}).
		Joins("JOIN likes ON likes.video_id = videos.id").
		Where("likes.account_id = ?", accountID).
		Order("likes.created_at desc").
		Find(&videos).Error
	if err != nil {
		return videos, err
	}
	return videos, nil
}
