package video

import (
	"crypto/rand"
	"encoding/hex"
	"feedsystem_video/backend/internal/middleware/jwt"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService *VideoService
	//accountService *account.AccountService
}

func NewVideoHandler(videoService *VideoService) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}
func (vh *VideoHandler) PublishVideo(c *gin.Context) {
	var req PulishVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	username, err := jwt.GetAccountUsername(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	video := &Video{
		AuthorID:    authorID,
		Username:    username,
		Title:       req.Title,
		PlayURL:     req.PlayURL,
		CoverURL:    req.CoverURL,
		Description: req.Description,
		Createtime:  time.Now(),
	}
	if err := vh.videoService.Publish(c.Request.Context(), video); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, video)
}
func (vh *VideoHandler) DelVideo(c *gin.Context) {
	var req DelVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := vh.videoService.Delete(c.Request.Context(), req.ID, authorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "del video success"})
}
func (vh *VideoHandler) ListByAuthorID(c *gin.Context) {
	var req ListByAuthorIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videos, err := vh.videoService.ListByAuthorID(c.Request.Context(), req.AuthorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videos)
}
func (vh *VideoHandler) GetDetail(c *gin.Context) {
	var req GetDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	videos, err := vh.videoService.GetDetail(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videos)
}
func (vh *VideoHandler) UpdateLikesCount(c *gin.Context) {
	var req UpdateLikesCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := vh.videoService.UpdateLikesCount(c.Request.Context(), req.ID, req.LikeCount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "UpdateLikesCount success"})
}
func (vh *VideoHandler) UploadVideo(c *gin.Context) {
	authorID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"messsagr": "missing file"})
		return
	}
	const maxSize = 200 << 20
	if f.Size <= 0 || f.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file size"})
		return
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if ext != ".mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .mp4 is allowed"})
		return
	}
	date := time.Now().Format("20060102")
	relDir := filepath.Join("videos", fmt.Sprintf("%d", authorID), date)
	root := filepath.Join(".run", "uploads")
	absDir := filepath.Join(root, relDir)
	filename := randHex(16) + ext
	absPath := filepath.Join(absDir, filename)
	if err := c.SaveUploadedFile(f, absPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	urlPath := path.Join("/static", "videos", fmt.Sprintf("%d", authorID), date, filename)
	c.JSON(http.StatusOK, gin.H{
		"url":       buildAbsoluteURL(c, urlPath),
		"cover_url": buildAbsoluteURL(c, urlPath),
	})
}
func (vh *VideoHandler) UploadCover(c *gin.Context) {
	authorID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := c.FormFile("filename")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	const MaxSize = 10 << 20
	if f.Size <= 0 || f.Size > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file size"})
		return
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .jpg/.jpeg/.png/.webp is allowed"})
		return
	}
	date := time.Now().Format("20060102")
	relDir := filepath.Join("covers", fmt.Sprintf("%d", authorID), date)
	root := filepath.Join(".run", "uploads")
	absDir := filepath.Join(root, relDir)
	if err := os.Mkdir(absDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := randHex(16)
	absPath := filepath.Join(absDir, filename)
	if err := c.SaveUploadedFile(f, absPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	urlPath := path.Join("/static", "covers", fmt.Sprintf("%d", authorID), date, filename)
	c.JSON(http.StatusOK, gin.H{
		"url":       buildAbsoluteURL(c, urlPath),
		"cover_url": buildAbsoluteURL(c, urlPath),
	})
}

func buildAbsoluteURL(c *gin.Context, p string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if xf := c.GetHeader("X-Forwarded-Proto"); xf != "" {
		scheme = xf
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, p)
}
func randHex(num int64) string {
	b := make([]byte, num)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
