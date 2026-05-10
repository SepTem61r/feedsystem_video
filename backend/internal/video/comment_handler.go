package video

import (
	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/middleware/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService    *CommentService
	accountRepository *account.AccountRepository
}

func NewCommentHandler(commentService *CommentService, accountRepository *account.AccountRepository) *CommentHandler {
	return &CommentHandler{
		commentService:    commentService,
		accountRepository: accountRepository,
	}
}
func (ch *CommentHandler) PublishComment(c *gin.Context) {
	var req CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VideoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id need > 0"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is require"})
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
	comment := &Comment{
		Username: username,
		VideoID:  req.VideoID,
		AuthorID: authorID,
		Content:  req.Content,
	}
	if err := ch.commentService.Publish(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "publish comment success"})
}

func (ch *CommentHandler) DeleteComment(c *gin.Context) {
	var req CommentDelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CommentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_id is required"})
		return
	}
	accountID, err := jwt.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ch.commentService.Delete(c.Request.Context(), req.CommentID, accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "delete comment success"})
}

func (ch *CommentHandler) GetAllComment(c *gin.Context) {
	var req CommentGetAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VideoID == 0 {
		c.JSON(400, gin.H{"error": "video_id is required"})
		return
	}
	comments, err := ch.commentService.GetAll(c.Request.Context(), req.VideoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if comments == nil {
		comments = []Comment{}
	}
	c.JSON(http.StatusOK, comments)
}
