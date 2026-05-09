package main

import (
	"fmt"
	"log"

	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/config"
	"feedsystem_video/backend/internal/db"
	"feedsystem_video/backend/internal/middleware/jwt"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"

	"github.com/gin-gonic/gin"
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

	rabbitClient, err := rabbitmq.NewFromEnv(&cfg.RabbitMQ)
	if err != nil {
		log.Printf("warning: failed to connect to rabbitmq: %v", err)
	} else {
		defer rabbitClient.Close()
	}

	accountRepo := account.NewAccountRepository(database)
	accountSvc := account.NewAccountService(accountRepo, redisClient)
	accountHandler := account.NewAccountHandler(accountSvc)

	r := gin.Default()

	r.POST("/api/account/register", accountHandler.CreateAccount)
	r.POST("/api/account/login", accountHandler.Login)

	auth := r.Group("/api/account")
	auth.Use(jwt.JWTAuth(accountRepo, redisClient))
	{
		auth.POST("/rename", accountHandler.Rename)
		auth.POST("/change-password", accountHandler.ChangePassword)
		auth.POST("/logout", accountHandler.Logout)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
