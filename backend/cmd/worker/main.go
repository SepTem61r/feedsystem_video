package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/config"
	"feedsystem_video/backend/internal/db"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"feedsystem_video/backend/internal/social"
	"feedsystem_video/backend/internal/video"
	"feedsystem_video/backend/worker"
)

func main() {
	cfg, usedDefault, err := config.LoadLocalDev("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if usedDefault {
		log.Println("using default local config")
	}

	database, err := db.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.CloseDB(database)

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	redisClient, err := rediscache.NewFromEnv(&cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	rabbitClient, err := rabbitmq.NewRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		log.Printf("warning: rabbitmq unavailable, workers will not start: %v", err)
		log.Println("api server handles writes via direct db fallback")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		return
	}
	defer rabbitClient.Close()

	// Declare exchanges/queues once via the main channel
	_, _ = rabbitmq.NewLikesCountMQ(rabbitClient)
	_, _ = rabbitmq.NewPopularityMQ(rabbitClient)
	_, _ = rabbitmq.NewCommentMQ(rabbitClient)
	_, _ = rabbitmq.NewSocialMQ(rabbitClient)
	_, err = rabbitmq.NewTimelineMQ(rabbitClient)
	if err != nil {
		log.Fatalf("failed to create timeline mq: %v", err)
	}

	videoRepo := video.NewVideoRepository(database)
	likeRepo := video.NewLikeRepository(database)
	commentRepo := video.NewCommentRepository(database)
	socialRepo := social.NewSocialRepository(database)
	accountRepo := account.NewAccountRepository(database)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Each worker gets its own channel to avoid sharing a single channel across goroutines.
	likeCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open like worker channel: %v", err)
	}
	likeWorker := worker.NewLikeWork(likeCh, videoRepo, likeRepo, "like.events")
	go func() {
		log.Println("like worker started")
		if err := likeWorker.Run(ctx); err != nil {
			log.Printf("like worker stopped: %v", err)
		}
	}()

	popularityCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open popularity worker channel: %v", err)
	}
	popularityWorker := worker.NewPopularityWorker(popularityCh, redisClient, "video.popularity.cache.queue")
	go func() {
		log.Println("popularity worker started")
		if err := popularityWorker.Run(ctx); err != nil {
			log.Printf("popularity worker stopped: %v", err)
		}
	}()

	commentCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open comment worker channel: %v", err)
	}
	commentWorker := worker.NewCommentWorker(commentCh, videoRepo, commentRepo, "comment.events")
	go func() {
		log.Println("comment worker started")
		if err := commentWorker.Run(ctx); err != nil {
			log.Printf("comment worker stopped: %v", err)
		}
	}()

	socialCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open social worker channel: %v", err)
	}
	socialWorker := worker.NewSocialWorker(socialCh, socialRepo, accountRepo, "social.events")
	go func() {
		log.Println("social worker started")
		if err := socialWorker.Run(ctx); err != nil {
			log.Printf("social worker stopped: %v", err)
		}
	}()

	// Outbox publisher gets its own channel.
	outboxCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open outbox channel: %v", err)
	}
	outboxRabbit := &rabbitmq.RabbitMQ{Conn: rabbitClient.Conn, Ch: outboxCh}
	timelineMQ, err := rabbitmq.NewTimelineMQ(outboxRabbit)
	if err != nil {
		log.Fatalf("failed to create outbox timeline mq: %v", err)
	}
	worker.StartOutboxPoller(database, timelineMQ)

	// Timeline consumer gets its own channel.
	consumerCh, err := rabbitClient.Channel()
	if err != nil {
		log.Fatalf("failed to open consumer channel: %v", err)
	}
	consumerRabbit := &rabbitmq.RabbitMQ{Conn: rabbitClient.Conn, Ch: consumerCh}
	consumerTimelineMQ, err := rabbitmq.NewTimelineMQ(consumerRabbit)
	if err != nil {
		log.Fatalf("failed to create consumer timeline mq: %v", err)
	}
	worker.StartConsumer(consumerTimelineMQ, "video.timeline.update.queue", redisClient)

	log.Println("all workers started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down workers...")
	cancel()
}
