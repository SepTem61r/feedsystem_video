package social

import (
	"context"
	"feedsystem_video/backend/internal/account"

	"gorm.io/gorm"
)

type SocialRepository struct {
	db *gorm.DB
}

// NewSocialRepository 创建 SocialRepository 实例
func NewSocialRepository(db *gorm.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// Follow 插入一条关注记录
func (sr *SocialRepository) Follow(ctx context.Context, social *Social) error {
	return sr.db.WithContext(ctx).Create(social).Error
}

// Unfollow 按关注关系删除一条记录
func (sr *SocialRepository) Unfollow(ctx context.Context, social *Social) error {
	return sr.db.WithContext(ctx).
		Where("follower_id = ? AND vlogger_id = ?", social.FollowerID, social.VloggerID).
		Delete(&Social{}).Error
}

// GetAllFollowers 查询某博主的所有粉丝账号
func (sr *SocialRepository) GetAllFollowers(ctx context.Context, vloggerID uint) ([]*account.Account, error) {
	var relations []Social
	if err := sr.db.WithContext(ctx).Model(&Social{}).Where("vlogger_id = ?", vloggerID).Find(&relations).Error; err != nil {
		return nil, err
	}
	followersIDs := make([]uint, 0, len(relations))
	for _, rel := range relations {
		followersIDs = append(followersIDs, rel.FollowerID)
	}
	if len(followersIDs) == 0 {
		return []*account.Account{}, nil
	}
	var followers []*account.Account
	if err := sr.db.WithContext(ctx).Model(&account.Account{}).Where("id IN ?", followersIDs).Find(&followers).Error; err != nil {
		return nil, err
	}
	return followers, nil
}

// GetAllVloggers 查询某用户关注的所有博主账号
func (sr *SocialRepository) GetAllVloggers(ctx context.Context, folowerID uint) ([]*account.Account, error) {
	var relations []Social
	if err := sr.db.WithContext(ctx).Model(&Social{}).Where("follower_id = ?", folowerID).Find(&relations).Error; err != nil {
		return nil, err
	}
	vloggerIDs := make([]uint, 0, len(relations))
	for _, rel := range relations {
		vloggerIDs = append(vloggerIDs, rel.VloggerID)
	}
	if len(vloggerIDs) == 0 {
		return []*account.Account{}, nil
	}
	var vloggers []*account.Account
	if err := sr.db.WithContext(ctx).Where("id IN ?", vloggerIDs).Find(&vloggers).Error; err != nil {
		return nil, err
	}
	return vloggers, nil
}

// IsFollowed 判断是否存在关注关系
func (sr *SocialRepository) IsFollowed(ctx context.Context, social *Social) (bool, error) {
	var count int64
	if err := sr.db.WithContext(ctx).Model(&Social{}).Where("follower_id = ? AND vlogger_id = ?", social.FollowerID, social.VloggerID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
