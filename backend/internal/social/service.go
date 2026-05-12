package social

import (
	"context"
	"errors"
	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
)

type SocialService struct {
	socialRepo  *SocialRepository
	accountRepo *account.AccountRepository
	socialMQ    *rabbitmq.SocialMQ
}

// NewSocialService 创建 SocialService 实例
func NewSocialService(socialRepo *SocialRepository, accountRepo *account.AccountRepository, socialMQ *rabbitmq.SocialMQ) *SocialService {
	return &SocialService{
		socialRepo:  socialRepo,
		accountRepo: accountRepo,
		socialMQ:    socialMQ,
	}
}

// Follow 关注（校验账号存在 + 防重复 + 消息队列/直接写入）
func (ss *SocialService) Follow(ctx context.Context, social *Social) error {
	_, err := ss.accountRepo.FindByID(ctx, social.VloggerID)
	if err != nil {
		return err
	}
	_, err = ss.accountRepo.FindByID(ctx, social.FollowerID)
	if err != nil {
		return err
	}
	if social.FollowerID == social.VloggerID {
		return errors.New("can not follow yourself")
	}
	isFollowed, err := ss.socialRepo.IsFollowed(ctx, social)
	if err != nil {
		return err
	}
	if isFollowed {
		return errors.New("already follow")
	}
	if ss.socialMQ != nil {
		if err := ss.socialMQ.Follow(ctx, social.FollowerID, social.VloggerID); err != nil {
			return err
		}
	}
	return ss.socialRepo.Follow(ctx, social)
}

// Unfollow 取消关注（校验账号存在 + 防未关注取关）
func (ss *SocialService) Unfollow(ctx context.Context, social *Social) error {
	_, err := ss.accountRepo.FindByID(ctx, social.VloggerID)
	if err != nil {
		return err
	}
	_, err = ss.accountRepo.FindByID(ctx, social.FollowerID)
	if err != nil {
		return err
	}
	if social.FollowerID == social.VloggerID {
		return errors.New("can not unfollow yourself")
	}
	isFollowed, err := ss.socialRepo.IsFollowed(ctx, social)
	if err != nil {
		return err
	}
	if !isFollowed {
		return errors.New("already unfollow")
	}
	return ss.socialRepo.Unfollow(ctx, social)
}

// GetAllFollowers 获取所有粉丝
func (ss *SocialService) GetAllFollowers(ctx context.Context, vloggerID uint) ([]*account.Account, error) {
	_, err := ss.accountRepo.FindByID(ctx, vloggerID)
	if err != nil {
		return nil, err
	}
	return ss.socialRepo.GetAllFollowers(ctx, vloggerID)
}

// GetAllVloggers 获取所有关注
func (ss *SocialService) GetAllVloggers(ctx context.Context, followerID uint) ([]*account.Account, error) {
	_, err := ss.accountRepo.FindByID(ctx, followerID)
	if err != nil {
		return nil, err
	}
	return ss.socialRepo.GetAllVloggers(ctx, followerID)
}
