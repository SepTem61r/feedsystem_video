package worker

import (
	"context"
	"encoding/json"
	"errors"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	"feedsystem_video/backend/internal/video"
	"log"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CommentWorker struct {
	ch          *amqp.Channel
	videoRepo   *video.VideoRepository
	commentRepo *video.CommentRepository
	queue       string
}

func NewCommentWorker(ch *amqp.Channel, videoRepo *video.VideoRepository, commentRepo *video.CommentRepository, queue string) *CommentWorker {
	return &CommentWorker{
		ch:          ch,
		videoRepo:   videoRepo,
		commentRepo: commentRepo,
		queue:       queue,
	}
}
func (cw *CommentWorker) Run(ctx context.Context) error {
	if cw == nil || cw.ch == nil || cw.commentRepo == nil || cw.videoRepo == nil {
		return errors.New("comment worker is not initialized")
	}
	if cw.queue == "" {
		return errors.New("queue is required")
	}

	deliveries, err := cw.ch.Consume(
		cw.queue,
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
			cw.handleDelivery(ctx, d)
		}
	}
}
func (cw *CommentWorker) handleDelivery(ctx context.Context, d amqp.Delivery) {
	if err := cw.pocess(ctx, d.Body); err != nil {
		log.Printf("comment worker: failed to process message: %v", err)
		_ = d.Nack(false, true)
		return
	}
	_ = d.Ack(false)
}
func (cw *CommentWorker) pocess(ctx context.Context, body []byte) error {
	var evt rabbitmq.CommentEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}

	switch evt.Action {
	case "publish":
		return cw.ApplyPublish(ctx, &evt)
	case "delete":
		return cw.ApplyDelete(ctx, &evt)
	default:
		return nil
	}
}
func (cw *CommentWorker) ApplyPublish(ctx context.Context, evt *rabbitmq.CommentEvent) error {
	if evt == nil || evt.VideoID == 0 || evt.AuthorID == 0 || strings.TrimSpace(evt.Content) == "" {
		return nil
	}
	ok, err := cw.videoRepo.IsExist(ctx, evt.VideoID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	c := &video.Comment{
		Username: strings.TrimSpace(evt.Username),
		VideoID:  evt.VideoID,
		AuthorID: evt.AuthorID,
		Content:  strings.TrimSpace(evt.Content),
	}
	if err := cw.commentRepo.CreateComment(ctx, c); err != nil {
		return err
	}
	return cw.videoRepo.ChangePopularity(ctx, evt.VideoID, 1)
}
func (cw *CommentWorker) ApplyDelete(ctx context.Context, evt *rabbitmq.CommentEvent) error {
	if evt == nil || evt.VideoID == 0 || evt.AuthorID == 0 || strings.TrimSpace(evt.Content) == "" {
		return nil
	}
	ok, err := cw.videoRepo.IsExist(ctx, evt.VideoID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	c := &video.Comment{
		Username: strings.TrimSpace(evt.Username),
		VideoID:  evt.VideoID,
		AuthorID: evt.AuthorID,
		Content:  strings.TrimSpace(evt.Content),
	}
	if err := cw.commentRepo.DeleteComment(ctx, c); err != nil {
		return err
	}
	return cw.videoRepo.ChangePopularity(ctx, evt.VideoID, -1)
}
