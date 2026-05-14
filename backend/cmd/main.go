package main

import (
	"fmt"
	"log"

	"feedsystem_video/backend/internal/config"
	"feedsystem_video/backend/internal/db"
	internalhttp "feedsystem_video/backend/internal/http"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
)

func main() {
	cfg, usedDefault, err := config.LoadLocalDev("./backend/configs/config.yaml")
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
		log.Printf("warning: failed to connect to rabbitmq: %v", err)
	} else {
		defer rabbitClient.Close()
	}

	r := internalhttp.SetRouter(database, redisClient, rabbitClient)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
