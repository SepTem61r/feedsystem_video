package worker

import (
	"context"
	"encoding/json"
	"errors"
	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	"feedsystem_video/backend/internal/social"
	"log"

	"github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SocialWorker struct {
	ch          *amqp.Channel
	socialRepo  *social.SocialRepository
	accountRepo *account.AccountRepository
	queue       string
}

func NewSocialWorker(ch *amqp.Channel, socialRepo *social.SocialRepository, accountRepo *account.AccountRepository, queue string) *SocialWorker {
	return &SocialWorker{
		ch:          ch,
		socialRepo:  socialRepo,
		accountRepo: accountRepo,
		queue:       queue,
	}
}
func (sw *SocialWorker) Run(ctx context.Context) error {
	if sw == nil || sw.ch == nil || sw.accountRepo == nil || sw.socialRepo == nil {
		return errors.New("social work is not initialized")
	}
	if sw.queue == "" {
		return errors.New("queue is required")
	}
	deliverys, err := sw.ch.Consume(
		sw.queue,
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
		case d, ok := <-deliverys:
			if !ok {
				return errors.New("deliveries channel closed")
			}
			sw.handleDelivery(ctx, d)
		}
	}
}
func (sw *SocialWorker) handleDelivery(ctx context.Context, d amqp.Delivery) {
	if err := sw.process(ctx, d.Body); err != nil {
		log.Printf("social worker: failed to process message: %v", err)
		_ = d.Nack(false, true)
		return
	}
	_ = d.Ack(false)
}
func (sw *SocialWorker) process(ctx context.Context, body []byte) error {
	var evt rabbitmq.SocialEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		// 解析事件失败，直接丢弃
		return nil
	}
	if evt.FollowerID == 0 || evt.VloggerID == 0 {
		return nil
	}
	switch evt.Action {
	case "follow":
		err := sw.socialRepo.Follow(ctx, &social.Social{
			FollowerID: evt.FollowerID,
			VloggerID:  evt.VloggerID,
		})
		if err == nil {
			return nil
		}
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil
		}
		return err
	case "unfollow":
		return sw.socialRepo.Unfollow(ctx, &social.Social{
			FollowerID: evt.FollowerID,
			VloggerID:  evt.VloggerID,
		})
	default:
		return nil
	}
}
