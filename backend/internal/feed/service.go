package feed

import (
	"context"
	"encoding/json"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"feedsystem_video/backend/internal/video"
	"fmt"
	"log"
	"sync"

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
	if len(videoIDs) == 0 || videoIDs == nil {
		return []*video.Video{}, nil
	}
	videoMap := make(map[uint]*video.Video)
	//L1
	var missedL1 []uint
	for _, id := range videoIDs {
		cacheKey := fmt.Sprintf("video:entiy:%d", id)
		fs.localcache.Get(cacheKey)
		if fs.localcache != nil {
			if v, found := fs.localcache.Get(cacheKey); found {
				if data, ok := v.(video.Video); ok {
					videoMap[id] = &data
					continue
				}
			}
			missedL1 = append(missedL1, id)
		}
	}
	if len(missedL1) == 0 {
		return buildOrderedResult(videoIDs, videoMap), nil
	}

	//L2'redis

	var missedL2 []uint
	if len(missedL1) > 0 {
		cacheKeys := make([]string, len(missedL1))
		for _, id := range missedL1 {
			cacheKeys[id] = fmt.Sprintf("video:entiy:%d", id)
		}
		cacheCtx, cacnel := context.WithTimeout(ctx, 50*time.Millisecond)
		result, err := fs.rediscache.MGet(cacheCtx, cacheKeys...)
		cacnel()
		if err == nil {
			for i, res := range result {
				id := missedL1[i]
				if res != nil {
					if str, ok := res.(string); ok {
						var v video.Video
						if err := json.Unmarshal([]byte(str), &v); err == nil {
							videoMap[id] = &v
							if fs.localcache != nil {
								//写入本地缓存
								fs.localcache.Set(cacheKeys[id], v, time.Hour)
							}
							continue
						}
					}
				}
				missedL2 = append(missedL2, id)
			}
		} else {
			missedL2 = missedL1
			log.Fatal("L2 redis缓存失败，降级L3")
		}
	}
	if len(missedL2) == 0 {
		return buildOrderedResult(videoIDs, videoMap), nil
	}
	//L3
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, id := range missedL2 {
		wg.Add(1)
		go func(videoID uint) {
			defer wg.Done()
			sfkey := fmt.Sprintf("sf:entiy:%d", videoID)
			v, err, _ := fs.requestGroup.Do(sfkey, func() (interface{}, error) {
				videoList, err := fs.repo.GetByIDs(ctx, []uint{videoID})
				if videoList == nil || len(videoList) == 0 {
					return nil, err
				}
				saftCopy := *videoList[0]
				cacheKey := fmt.Sprintf("video:entiy:%d", saftCopy.ID)
				if b, err := json.Marshal(saftCopy); err != nil {
					go func(k string, b []byte) {
						setCtx, setcel := context.WithTimeout(context.Background(), 50*time.Millisecond)
						defer setcel()
						fs.rediscache.SetBytes(setCtx, k, b, time.Hour)
					}(cacheKey, b)

				}
				return *videoList[0], err
			})
			//写回本地缓存
			if err == nil && v != nil {
				if fs.localcache != nil {
					safeCopy := *(v.(*video.Video))
					mu.Lock()
					videoMap[id] = &safeCopy
					mu.Unlock()
					fs.localcache.Set(fmt.Sprintf("video:entiy:%d", safeCopy.ID), safeCopy, time.Hour)
				}
			}
		}(id)
	}
	wg.Wait()
	return buildOrderedResult(videoIDs, videoMap), nil
}
func (fs *FeedService) ListLatest(ctx context.Context, limit int64, latestBefore time.Time) error {
	return nil
}
