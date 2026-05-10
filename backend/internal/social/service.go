package social

import (
	"feedsystem_video/backend/internal/account"
)

type SocialService struct {
	socialRepository  *SocialRepository
	accountRepository *account.AccountRepository
}
