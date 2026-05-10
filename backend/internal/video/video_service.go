package video

import (
	"context"
	"encoding/json"
	"errors"
	"feedsystem_video/backend/internal/middleware/rabbitmq"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type VideoService struct {
	repo      *VideoRepository
	cache     *rediscache.Client
	cacheTTL  time.Duration
	popularMQ *rabbitmq.PopularityMQ
}

// NewVideoService 创建 VideoService 实例
func NewVideoService(repo *VideoRepository, cache *rediscache.Client, popularMQ *rabbitmq.PopularityMQ) *VideoService {
	return &VideoService{repo: repo, cache: cache, popularMQ: popularMQ}
}

// Publish 发布视频（校验字段 + 事务写入 video 和 outbox 消息）
func (vs *VideoService) Publish(ctx context.Context, video *Video) error {
	if video == nil {
		return errors.New("video null")
	}
	video.Title = strings.TrimSpace(video.Title)
	video.PlayURL = strings.TrimSpace(video.PlayURL)
	video.CoverURL = strings.TrimSpace(video.CoverURL)
	if video.Title == "" {
		return errors.New("title is require")
	}
	if video.PlayURL == "" {
		return errors.New("play url is require")
	}
	if video.CoverURL == "" {
		return errors.New("cover url is require")
	}
	err := vs.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&video).Error; err != nil {
			return err
		}
		msg := OutboxMsg{
			VideoID:    video.ID,
			EventType:  "video_published",
			Status:     "pending",
			CreateTime: video.Createtime,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

// Delete 删除视频（校验作者身份，同时清理缓存）
func (vs *VideoService) Delete(ctx context.Context, id uint, authorID uint) error {
	video, err := vs.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if video == nil {
		return errors.New("video not found")
	}
	if video.AuthorID != authorID {
		return errors.New("unauthorized")
	}
	if err := vs.repo.DelVideo(ctx, id); err != nil {
		return err
	}
	if vs.cache != nil {
		cachekey := fmt.Sprintf("video:detail:id = %d", id)
		_ = vs.cache.Del(context.Background(), cachekey)
	}
	return nil
}

// ListByAuthorID 按作者ID查询视频列表
func (vs *VideoService) ListByAuthorID(ctx context.Context, authorID uint) ([]Video, error) {
	videos, err := vs.repo.ListByAuthorID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// GetDetail 获取视频详情（命中缓存返回，未命中时加锁查库回填缓存，带自旋等待）
func (vs *VideoService) GetDetail(ctx context.Context, id uint) (*Video, error) {
	cachekey := fmt.Sprintf("video:detail:id=%d", id)

	getCache := func() (*Video, bool, bool) {
		opCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		b, err := vs.cache.GetBytes(opCtx, cachekey)
		if err != nil {
			return nil, false, rediscache.IsMiss(err)
		}
		var cached Video
		if err := json.Unmarshal(b, &cached); err != nil {
			return nil, false, false
		}
		return &cached, true, false
	}
	setCache := func(video *Video) {
		b, err := json.Marshal(video)
		if err != nil {
			return
		}
		opCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		_ = vs.cache.SetBytes(opCtx, cachekey, b, vs.cacheTTL)

	}
	if vs.cache != nil {
		//缓存命中
		if v, ok, miss := getCache(); ok {
			return v, nil
		} else if miss {
			lockKey := "lock:" + cachekey
			lockCtx, lockCancel := context.WithTimeout(ctx, 50*time.Millisecond)
			token, locked, lockErr := vs.cache.Lock(lockCtx, lockKey, vs.cacheTTL)
			lockCancel()
			if lockErr == nil && locked {
				defer func() { _ = vs.cache.UnLock(context.Background(), lockKey, token) }()
				if v, ok, _ := getCache(); ok {
					return v, nil
				}
				video, err := vs.repo.GetByID(ctx, id)
				if err != nil {
					return nil, err
				}
				setCache(video)
				return video, nil
			}
			//没拿到锁，自旋等待
			for i := 0; i < 5; i++ {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(20 * time.Millisecond):
				}
				if v, ok, _ := getCache(); ok {
					return v, nil
				}
			}
		}
	}
	//无缓存 / 所有缓存路径都失败
	video, err := vs.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vs.cache != nil {
		setCache(video)
	}
	return video, nil
}

// UpdateLikesCount 更新视频点赞数

func (vs *VideoService) UpdateLikesCount(ctx context.Context, id uint, likesCount int64) error {
	if err := vs.repo.UpdateLikeCount(ctx, id, likesCount); err != nil {
		return err
	}
	return nil
}
