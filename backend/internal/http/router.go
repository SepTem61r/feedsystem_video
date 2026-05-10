package http

import (
	"feedsystem_video/backend/internal/account"
	"feedsystem_video/backend/internal/middleware/jwt"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	"feedsystem_video/backend/internal/middleware/ratelimit"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"feedsystem_video/backend/internal/video"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetRouter(db *gorm.DB, cache *rediscache.Client, rmq *rabbitmq.RabbitMQ) *gin.Engine {
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("SetTrustedProxies failed: %v", err)
	}
	r.Static("/static", "./.run/uploads")
	// rate_limit
	// rate_limit
	loginLimiter := ratelimit.Limit(cache, "account_login", 10, time.Minute, ratelimit.KeyByIP)
	registerLimiter := ratelimit.Limit(cache, "account_register", 5, time.Hour, ratelimit.KeyByIP)

	likeLimiter := ratelimit.Limit(cache, "like_write", 30, time.Minute, ratelimit.KeyByAccount)
	commentLimiter := ratelimit.Limit(cache, "comment_write", 10, time.Minute, ratelimit.KeyByAccount)
	//socialLimiter := ratelimit.Limit(cache, "social_write", 20, time.Minute, ratelimit.KeyByAccount)

	//account
	accountRepository := account.NewAccountRepository(db)
	accountService := account.NewAccountService(accountRepository, cache)
	accountHandler := account.NewAccountHandler(accountService)
	accountGroup := r.Group("/account")
	{
		accountGroup.POST("/register", registerLimiter, accountHandler.CreateAccount)
		accountGroup.POST("/login", loginLimiter, accountHandler.Login)
		accountGroup.POST("/changePassword", accountHandler.ChangePassword)
		accountGroup.POST("/findByID", accountHandler.FindByID)
		accountGroup.POST("/findByUsername", accountHandler.FindByUsername)
	}
	protectedAccountGroup := accountGroup.Group("")
	protectedAccountGroup.Use(jwt.JWTAuth(accountRepository, cache))
	{
		protectedAccountGroup.POST("/logout", accountHandler.Logout)
		protectedAccountGroup.POST("/rename", accountHandler.Rename)

	}

	//video
	popularityMQ, err := rabbitmq.NewPopularityMQ(rmq)
	if err != nil {
		log.Printf("PopularityMQ init failed (mq disabled): %v", err)
		popularityMQ = nil
	}
	videoRepository := video.NewVideoRepository(db)
	videoService := video.NewVideoService(videoRepository, cache, popularityMQ)
	videoHandler := video.NewVideoHandler(videoService)
	videoGroup := r.Group("/video")
	{
		videoGroup.POST("/listByButhorID", videoHandler.ListByAuthorID)
		videoGroup.POST("/getDetail", videoHandler.GetDetail)

	}
	protectedVideoGroup := videoGroup.Group("")
	protectedVideoGroup.Use(jwt.JWTAuth(accountRepository, cache))
	{
		protectedVideoGroup.POST("/publish", videoHandler.PublishVideo)
		protectedVideoGroup.POST("/uploadVideo", videoHandler.UploadVideo)
		protectedVideoGroup.POST("/uploadCover", videoHandler.UploadCover)
		protectedVideoGroup.POST("/delete", videoHandler.DelVideo)
	}

	//like
	likeMQ, err := rabbitmq.NewLikesCountMQ(rmq)
	if err != nil {
		log.Printf("likeMQ init failed (mq disabled): %v", err)
		likeMQ = nil
	}
	likeRepository := video.NewLikeRepository(db)
	likeService := video.NewLikeService(likeRepository, videoRepository, cache, likeMQ, popularityMQ)
	likeHandler := video.NewLikeHandler(likeService)
	likeGroup := r.Group("/like")
	protectedLikeGroup := likeGroup.Group("")
	protectedLikeGroup.Use(jwt.JWTAuth(accountRepository, cache))
	{
		protectedLikeGroup.POST("/like", likeLimiter, likeHandler.Like)
		protectedLikeGroup.POST("/unlike", likeLimiter, likeHandler.Unlike)
		protectedLikeGroup.POST("/isLiked", likeHandler.IsLiked)
		protectedLikeGroup.POST("/listMyLikedVideos", likeHandler.ListMyLikedVideos)
	}

	//comment
	commentMQ, err := rabbitmq.NewCommentMQ(rmq)
	if err != nil {
		log.Printf("commentMQ init failed (mq disabled): %v", err)
		commentMQ = nil
	}
	commentRepository := video.NewCommentRepository(db)
	commentService := video.NewCommentService(commentRepository, cache, popularityMQ, videoRepository, commentMQ)
	commentHandler := video.NewCommentHandler(commentService, accountRepository)
	commentGroup := r.Group("/comment")
	{
		commentGroup.POST("/getAll", commentHandler.GetAllComment)
	}
	protectedCommentGroup := commentGroup.Group("")
	protectedCommentGroup.Use(jwt.JWTAuth(accountRepository, cache))
	{
		protectedCommentGroup.POST("/publish", commentLimiter, commentHandler.PublishComment)
		protectedCommentGroup.POST("/delete", commentLimiter, commentHandler.DeleteComment)
	}

	return r
}
