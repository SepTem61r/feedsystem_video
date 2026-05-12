package feed

import (
	"context"
	"encoding/json"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"feedsystem_video/backend/internal/video"
	"fmt"
	"log"

	"github.com/patrickmn/go-cache"

	"golang.org/x/sync/singleflight"
	"time"
)

type FeedService struct {
	repo         *FeedRepository
	likeRepo     *video.LikeRepository
	rediscache   *rediscache.Client
	localcache   *cache.Cache
	cacheTTL     time.Duration
	requestGroup singleflight.Group
}

func NewFeedService(repo *FeedRepository, likeRepo *video.LikeRepository, rediscache *rediscache.Client) *FeedService {
	return &FeedService{
		repo:       repo,
		likeRepo:   likeRepo,
		rediscache: rediscache,
		localcache: cache.New(3*time.Second, 5*time.Second),
		cacheTTL:   24 * time.Hour,
	}
}
func buildOrderedResult(orderedIDs []uint, dataMap map[uint]*video.Video) []*video.Video {
	res := make([]*video.Video, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		if v, exits := dataMap[id]; exits && v != nil {
			res = append(res, v)
		}
	}
	return res
}
func (fs *FeedService) GetVideoByIDs(ctx context.Context, videoIDs []uint) ([]*video.Video, error) {
	// GetVideoByIDs 批量获取视频信息
	// 采用 L1(本地缓存) -> L2(Redis) -> L3(MySQL) 三级架构
	if len(videoIDs) == 0 {
		return []*video.Video{}, nil
	}
	videomap := make(map[uint]*video.Video)
	//l1
	var missedL1 []uint
	for _, id := range videoIDs {
		cacheKey := fmt.Sprintf("video:entiy: %d", id)
		if fs.localcache != nil {
			if v, found := fs.localcache.Get(cacheKey); found {
				if data, ok := v.(video.Video); ok {
					videomap[id] = &data
					continue
				}
			}
		}
		missedL1 = append(missedL1, id)
	}
	if len(missedL1) == 0 {
		return buildOrderedResult(videoIDs, videomap), nil
	}
	//l2
	var missedL2 []uint
	if len(missedL1) > 0 {
		cacheKeys := make([]string, len(missedL1))
		for i, id := range missedL1 {
			cacheKeys[i] = fmt.Sprintf("video:entity:%d", id)
		}

		cacheCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		results, err := fs.rediscache.MGet(cacheCtx, cacheKeys...)
		cancel()

		if err == nil {
			for i, res := range results {
				id := missedL1[i]
				if res != nil {
					if str, ok := res.(string); ok {
						var v video.Video
						if err := json.Unmarshal([]byte(str), &v); err == nil {
							videomap[id] = &v
							// 回写更新 L1 本地缓存
							if fs.localcache != nil {
								fs.localcache.Set(cacheKeys[i], v, 5*time.Second)
							}
							continue
						}
					}
				}
				missedL2 = append(missedL2, id)
			}
		} else {
			// 如果 Redis 挂了或者超时了，全部降级到 L3
			missedL2 = missedL1
			log.Printf("L2 Redis MGet 失败，全部降级到 MySQL: %v", err)
		}
	}

	return buildOrderedResult(videoIDs, videomap), nil
}
func (fs *FeedService) ListLatest(ctx context.Context, limit int64, latestBefore time.Time) error {
	return nil
}
