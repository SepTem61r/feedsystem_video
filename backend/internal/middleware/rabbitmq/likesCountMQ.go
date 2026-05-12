package rabbitmq

import (
	"context"
	"errors"
	"time"
)

type LikeMQ struct {
	*RabbitMQ
}

const (
	likeExchange   = "like.events"
	likeQueue      = "like.events"
	likeBindingKey = "like.*"
	LikeRK         = "like.like"
	UnlikeRK       = "like.unlike"
)

type LikesCountEvent struct {
	EventID    string    `json:"event_id"`
	Action     string    `json:"action"`
	UserID     uint      `json:"user_id"`
	VideoID    uint      `json:"video_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func NewLikesCountMQ(base *RabbitMQ) (*LikeMQ, error) {
	if base == nil {
		return nil, errors.New("rabbitMQ base is null")
	}
	if err := base.DeclareTopic(likeExchange, likeQueue, likeBindingKey); err != nil {
		return nil, err
	}
	return &LikeMQ{RabbitMQ: base}, nil
}
func (l *LikeMQ) Like(ctx context.Context, userID, videoID uint) error {
	return l.Publish(ctx, "like", LikeRK, userID, videoID)

}
func (l *LikeMQ) UnLike(ctx context.Context, userID, videoID uint) error {
	return l.Publish(ctx, "unlike", UnlikeRK, userID, videoID)

}
func (l *LikeMQ) Publish(ctx context.Context, action, routingKey string, userID, videoID uint) error {
	if l == nil || l.RabbitMQ == nil {
		return errors.New("like mq is not initialized")
	}
	if userID == 0 || videoID == 0 {
		return errors.New("userID videoID is require")
	}
	id, err := newEventID(16)
	if err != nil {
		return err
	}
	event := LikesCountEvent{
		EventID:    id,
		Action:     action,
		UserID:     userID,
		VideoID:    videoID,
		OccurredAt: time.Now(),
	}
	return l.PublishJSON(ctx, likeExchange, routingKey, event)
}
