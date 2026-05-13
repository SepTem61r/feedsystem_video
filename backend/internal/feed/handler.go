package feed

import (
	"feedsystem_video/backend/internal/middleware/jwt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	feedService *FeedService
}

func NewFeedHandler(feedService *FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// nonNilFeedVideoItems 确保返回非nil切片（nil转为空切片）
func nonNilFeedVideoItems(items []FeedVideoItem) []FeedVideoItem {
	if items == nil {
		return []FeedVideoItem{}
	}
	return items
}

// 按照时间列表排序
func (fh *FeedHandler) ListLatest(c *gin.Context) {
	var req ListLatestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 10
	}
	var latestTime time.Time
	if req.LatestTime > 0 {
		latestTime = time.UnixMilli(req.LatestTime)
	}
	viewerAccountID, err := jwt.GetAccountID(c)
	if err != nil {
		viewerAccountID = 0
	}
	feedItems, err := fh.feedService.ListLatest(c.Request.Context(), req.Limit, latestTime, viewerAccountID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	feedItems.VideoList = nonNilFeedVideoItems(feedItems.VideoList)
	c.JSON(http.StatusOK, feedItems)
}

// 按照点赞数排序(倒序)
func (fh *FeedHandler) ListLikesCount(c *gin.Context) {
	var req ListLikesCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 10
	}

	var cursor *LikesCountCursor
	if req.LikesCountBefore != nil || req.IDBefore != nil {
		if req.LikesCountBefore == nil || req.IDBefore == nil {
			c.JSON(400, gin.H{"error": "likes_count_before and id_before must be provided together"})
			return
		}
		likesCountBefore := *req.LikesCountBefore
		idBefore := *req.IDBefore

		if likesCountBefore < 0 {
			c.JSON(400, gin.H{"error": "invalid cursor: likes_count_before must be >= 0"})
			return
		}
		if idBefore == 0 {
			if likesCountBefore != 0 {
				c.JSON(400, gin.H{"error": "invalid cursor: id_before must be > 0"})
				return
			}
		} else {
			cursor = &LikesCountCursor{
				LikesCount: likesCountBefore,
				ID:         idBefore,
			}
		}
	}
	viewerAccountID, err := jwt.GetAccountID(c)
	if err != nil {
		viewerAccountID = 0
	}
	feedItems, err := fh.feedService.LikesCount(c.Request.Context(), req.Limit, cursor, viewerAccountID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	feedItems.VideoList = nonNilFeedVideoItems(feedItems.VideoList)
	c.JSON(http.StatusOK, feedItems)
}
func (fh *FeedHandler) ListByFollowing(c *gin.Context) {
	var req ListByFollowingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 10
	}
	viewerAccountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var latestTime time.Time
	if req.LatestTime > 0 {
		latestTime = time.Unix(req.LatestTime, 0)
	}
	feedItems, err := fh.feedService.ListByFollowing(c.Request.Context(), req.Limit, latestTime, viewerAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	feedItems.VideoList = nonNilFeedVideoItems(feedItems.VideoList)
	c.JSON(http.StatusOK, feedItems)

}
