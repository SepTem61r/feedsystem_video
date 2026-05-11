package worker

import (
	"context"
	"encoding/json"
	"errors"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	"feedsystem_video/backend/internal/video"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type LikeWorker struct {
	ch        *amqp.Channel
	videoRepo *video.VideoRepository
	likeRepo  *video.LikeRepository
	queue     string
}

func NewLikeWork(ch *amqp.Channel, videoRepo *video.VideoRepository, likeRepo *video.LikeRepository, queue string) *LikeWorker {
	return &LikeWorker{
		ch:        ch,
		videoRepo: videoRepo,
		likeRepo:  likeRepo,
		queue:     queue,
	}
}
func (lw *LikeWorker) Run(ctx context.Context) error {
	if lw == nil || lw.ch == nil || lw.likeRepo == nil || lw.videoRepo == nil {
		return errors.New(" like work is not initialized")
	}
	if lw.queue == "" {
		return errors.New("queue is require")
	}
	deliveries, err := lw.ch.Consume(
		lw.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("deliveries channel closed")
			}
			lw.handlerDelivery(ctx, d)
		}
	}
}
func (lw *LikeWorker) handlerDelivery(ctx context.Context, d amqp.Delivery) {
	if err := lw.process(ctx, d.Body); err != nil {
		log.Printf("like worker: failed to process message: %v", err)
		_ = d.Nack(false, true)
		return
	}
	_ = d.Ack(false)

}
func (lw *LikeWorker) process(ctx context.Context, body []byte) error {
	var evt rabbitmq.LikesCountEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return nil
	}
	if evt.UserID == 0 || evt.VideoID == 0 {
		return nil
	}

	switch evt.Action {
	case "like":
		return lw.applyLike(ctx, evt.UserID, evt.VideoID)
	case "unlike":
		return lw.applyUnlike(ctx, evt.UserID, evt.VideoID)
	default:
		return nil
	}
}
func (lw *LikeWorker) applyLike(ctx context.Context, userID, videoID uint) error {
	ok, err := lw.videoRepo.IsExist(ctx, videoID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	created, err := lw.likeRepo.LikeIgnoreDuplicate(ctx, &video.Like{
		VideoID:   videoID,
		AccountID: userID,
		CreatedAt: time.Time{},
	})
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	if err := lw.videoRepo.ChangeLikesCount(ctx, videoID, 1); err != nil {
		return err
	}
	return lw.videoRepo.ChangePopularity(ctx, videoID, 1)
}
func (lw *LikeWorker) applyUnlike(ctx context.Context, userID, videoID uint) error {
	ok, err := lw.videoRepo.IsExist(ctx, videoID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	created, err := lw.likeRepo.LikeIgnoreDuplicate(ctx, &video.Like{
		VideoID:   videoID,
		AccountID: userID,
		CreatedAt: time.Time{},
	})
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	if err := lw.videoRepo.ChangeLikesCount(ctx, videoID, -1); err != nil {
		return err
	}
	return lw.videoRepo.ChangePopularity(ctx, videoID, -1)
}
