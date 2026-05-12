package social

import (
	"feedsystem_video/backend/internal/middleware/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SocialHandler struct {
	socialService *SocialService
}

// NewSocialHandler 创建 SocialHandler 实例
func NewSocialHandler(socialService *SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

// Follow 关注
func (sh *SocialHandler) Follow(c *gin.Context) {
	var req FollowerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	followerID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	social := &Social{
		FollowerID: followerID,
		VloggerID:  req.VloggerID,
	}
	if err := sh.socialService.Follow(c.Request.Context(), social); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "follow success"})
}

// Unfollow 取消关注
func (sh *SocialHandler) Unfollow(c *gin.Context) {
	var req UnFollowerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	followerID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	social := &Social{
		FollowerID: followerID,
		VloggerID:  req.VloggerID,
	}
	if err := sh.socialService.Unfollow(c.Request.Context(), social); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unfollow success"})
}

// GetAllFollowers 获取所有粉丝
func (sh *SocialHandler) GetAllFollowers(c *gin.Context) {
	var req GetAllFollowersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vloggerID := req.VloggerID
	if vloggerID == 0 {
		accountID, err := jwt.GetAccountID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		vloggerID = accountID
	}
	followers, err := sh.socialService.GetAllFollowers(c.Request.Context(), vloggerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetAllFollowersResponse{Followers: followers})
}

// GetAllVloggers 获取所有关注
func (sh *SocialHandler) GetAllVloggers(c *gin.Context) {
	var req GetAllVloggersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	followerID := req.FollowerID
	if followerID == 0 {
		accountID, err := jwt.GetAccountID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		followerID = accountID
	}
	vloggers, err := sh.socialService.GetAllVloggers(c.Request.Context(), followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetAllVloggersResponse{Vloggers: vloggers})
}
