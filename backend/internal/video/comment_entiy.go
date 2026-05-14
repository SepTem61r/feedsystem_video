package video

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey", json:"id"`
	Username  string    `gorm:"index", json:"username"`
	VideoID   uint      `gorm:"index", json:"video_id"`
	AuthorID  uint      `gorm:"index", json:"author_id"`
	Content   string    `goem:"type:text", json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
type CommentRequest struct {
	VideoID uint   `json:"video_id"`
	Content string `json:"content"`
}
type CommentDelRequest struct {
	CommentID uint `json:"comment_id"`
}
type CommentGetAllRequest struct {
	VideoID uint `json:"video_id"`
}
