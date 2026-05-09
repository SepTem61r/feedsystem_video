package video

import (
	"feedsystem_video/backend/internal/middleware/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	likeService *LikeService
}

func NewLikeHandler(likeService *LikeService) *LikeHandler {
	return &LikeHandler{likeService: likeService}
}
func (lh *LikeHandler) Like(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VideoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
		return
	}
	accountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	like := &Like{VideoID: req.VideoID, AccountID: accountID}
	if err := lh.likeService.Like(c.Request.Context(), like); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "like success"})
}

func (lh *LikeHandler) Unlike(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VideoID <= 0 {
		c.JSON(400, gin.H{"error": "video_id is required"})
		return
	}
	accountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	like := &Like{VideoID: req.VideoID, AccountID: accountID}
	if err := lh.likeService.Unlike(c.Request.Context(), like); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unlike success"})
}
func (lh *LikeHandler) IsLiked(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VideoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
		return
	}

	accountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok, err := lh.likeService.Isliked(c.Request.Context(), req.VideoID, accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_liked": ok})
}
func (lh *LikeHandler) ListMyLikedVideos(c *gin.Context) {

	accountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	videos, err := lh.likeService.ListLikedVideos(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if videos == nil {
		videos = []Video{}
	}
	c.JSON(http.StatusOK, gin.H{"videos": videos})

}
