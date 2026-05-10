package social

import "gorm.io/gorm"

type SocialRepository struct {
	db *gorm.DB
}

func NewSocialRepository(db *gorm.DB) *SocialRepository {
	return &SocialRepository{db: db}
}
func (sr *SocialRepository) Follow(ctx context.Context, social *Social) error {
	return sr.db.WithContext(ctx).Create(social).Error
}
func (sr *SocialRepository) Unfollow(ctx context.Context, social *Social) error {
	return sr.db.WithContext(ctx).
		Where("follower_id = ? AND vlogger_id = ?", social.FollowerID, social.VloggerID).
		Delete(&Social{}).Error
}
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
func (sr *SocialRepository) GetAllVloggers(ctx context.Context, folowerID uint) ([]*account.Account, error) {
	var relations []Social
	if err := sr.db.WithContext(ctx).Model(&Social{}).Where("folower_id = ?", folowerID).Find(&relations).Error; err != nil {
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
