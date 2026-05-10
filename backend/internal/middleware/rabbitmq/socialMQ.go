package rabbitmq

import (
	"context"
	"errors"
	"time"
)

type SocialMQ struct {
	*RabbitMQ
}

const (
	socialExchange   = "social.events"
	socialQueue      = "social.events"
	socialBindingKey = "social.*"

	socialFollowRK   = "social.follow"
	socialUnfollowRK = "social.unfollow"
)

type SocialEvent struct {
	EventID    string    `json:"event_id"`
	Action     string    `json:"action"`
	FollowerID uint      `json:"follower_id"`
	VloggerID  uint      `json:"vlogger_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func NewSocialMQ(base *RabbitMQ) (*SocialMQ, error) {
	if base == nil {
		return nil, errors.New("rabbitMQ base is null")
	}
	if err := base.DeclareTopic(socialExchange, socialQueue, socialBindingKey); err != nil {
		return nil, err
	}
	return &SocialMQ{base}, nil
}

func (sMQ *SocialMQ) Follow(ctx context.Context, followerID, VloggerID uint) error {
	return sMQ.publish(ctx, "follow", socialFollowRK, followerID, VloggerID)
}
func (sMQ *SocialMQ) Unfollow(ctx context.Context, followerID, VloggerID uint) error {
	return sMQ.publish(ctx, "unfollow", socialUnfollowRK, followerID, VloggerID)
}
func (sMQ *SocialMQ) publish(ctx context.Context, action, routingKey string, followerID, vloggerID uint) error {
	if sMQ == nil || sMQ.RabbitMQ == nil {
		return errors.New("social mq is not initialized")
	}
	if followerID == 0 || vloggerID == 0 {
		return errors.New("follower_id and vlogger_id is require")
	}
	id, err := newEventID(16)
	if err != nil {
		return err
	}
	evt := &SocialEvent{
		EventID:    id,
		FollowerID: followerID,
		VloggerID:  vloggerID,
		OccurredAt: time.Time{},
	}
	return sMQ.PublishJSON(ctx, socialExchange, routingKey, evt)
}
