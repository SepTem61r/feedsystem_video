package feed

import (
	"context"
	"encoding/json"
	rediscache "feedsystem_video/backend/internal/middleware/redis"
	"feedsystem_video/backend/internal/video"
	"fmt"
	"log"
	"strconv"
	"sync"

	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"

	"golang.org/x/sync/singleflight"
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

// buildOrderedResult 按指定ID顺序从map中取出视频列表
func buildOrderedResult(orderedIDs []uint, dataMap map[uint]*video.Video) []*video.Video {
	res := make([]*video.Video, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		if v, exits := dataMap[id]; exits && v != nil {
			res = append(res, v)
		}
	}
	return res
}

// GetVideoByIDs 批量获取视频信息
func (fs *FeedService) GetVideoByIDs(ctx context.Context, videoIDs []uint) ([]*video.Video, error) {

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

// buildFeedVideos 构建FeedVideoItem列表并批量填充点赞状态
func (fs *FeedService) buildFeedVideos(ctx context.Context, videos []*video.Video, viewerAccountID uint) ([]FeedVideoItem, error) {
	feedVideos := make([]FeedVideoItem, 0, len(videos))
	videoIDs := make([]uint, len(videos))
	for i, v := range videos {
		videoIDs[i] = v.ID
	}
	likeMap, err := fs.likeRepo.BatchGetLiked(ctx, videoIDs, viewerAccountID)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		feedVideos = append(feedVideos, FeedVideoItem{
			ID:          video.ID,
			Author:      FeedAuthor{ID: video.AuthorID, Username: video.Username},
			Title:       video.Title,
			Description: video.Description,
			PlayURL:     video.PlayURL,
			CoverURL:    video.CoverURL,
			CreateTime:  video.Createtime.UnixMilli(),
			LikesCount:  video.LikesCount,
			IsLiked:     likeMap[video.ID],
		})
	}
	return feedVideos, nil
}

// 查询最新视频 (冷热分离 + 游标分页)
func (fs *FeedService) ListLatest(ctx context.Context, limit int, latestBefore time.Time, viewerAccountID uint) (ListLatestResponse, error) {
	zsetTail, err := fs.rediscache.ZRangeWithScores(ctx, "feed:global_timeline", 0, 0)
	if err != nil {
		return ListLatestResponse{}, err
	}
	iszsetEmpty := len(zsetTail) == 0
	//zset为空
	if iszsetEmpty {
		sfKey := "sf:fallback:global_timeline_rebuild"
		v, err, _ := fs.requestGroup.Do(sfKey, func() (interface{}, error) {
			dbvideos, err := fs.repo.ListLatest(ctx, 1000, time.Now())
			if err != nil {
				return nil, err
			}
			if dbvideos == nil || len(dbvideos) == 0 {
				return "DB_EMPTY", err
			}
			//重建zset
			bgCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			var zElements []redis.Z
			for _, vid := range dbvideos {
				zElements = append(zElements, redis.Z{
					Score:  float64(vid.Createtime.UnixMilli()),
					Member: fmt.Sprintf("%d", vid.ID),
				})
			}
			fs.rediscache.ZAdd(bgCtx, "feed:global_timeline", zElements...)
			return "SUCCESS ", nil
		})
		if err != nil {
			return ListLatestResponse{}, err
		}
		if v == "DB_EMPTY" {
			return ListLatestResponse{HasMore: false}, err
		}
		return fs.ListLatest(ctx, limit, latestBefore, viewerAccountID)
	}
	watermark := int64(zsetTail[0].Score)
	reqTime := time.Now().UnixMilli()
	if !latestBefore.IsZero() {
		reqTime = latestBefore.UnixMilli()
	}
	//冷数据查mysql
	var baseVideos []*video.Video

	if reqTime <= watermark {
		//冷数据降级查库
		sfKey := fmt.Sprintf("sf:cold:listLatest:%d:%d", limit, reqTime)
		v, err, _ := fs.requestGroup.Do(sfKey, func() (interface{}, error) {
			return fs.repo.ListLatest(ctx, limit, latestBefore)
		})
		if err != nil {
			return ListLatestResponse{}, err
		}
		baseVideos = v.([]*video.Video)
		// 不回写 ZSET，防止冷数据污染热点时间线

	} else {
		//热数据查redis
		maxScore := "+inf"
		if !latestBefore.IsZero() {
			maxScore = fmt.Sprintf("%d", reqTime-1) // 减 1 防重复（避免游标时间刚好命中某条视频）
		}
		videoIDsStr, err := fs.rediscache.ZRevRangeByScore(ctx, "feed:global_timeline", maxScore, "-inf", 0, int64(limit))
		if err != nil {
			return ListLatestResponse{}, err
		}
		var videoIDs []uint
		for _, idStr := range videoIDsStr {
			if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
				videoIDs = append(videoIDs, uint(id))
			}
		}
		if len(videoIDs) > 0 {
			baseVideos, err = fs.GetVideoByIDs(ctx, videoIDs)
			if err != nil {
				return ListLatestResponse{}, err
			}
		}
		//热数据不够 → 冷热拼接
		if len(baseVideos) < limit {
			remainLimit := limit - len(baseVideos) // 计算还差几个
			var coldCursor time.Time
			if len(baseVideos) > 0 {
				coldCursor = baseVideos[len(baseVideos)-1].Createtime
			} else {
				coldCursor = latestBefore
			}
			sfKey := fmt.Sprintf("sf:stitch:listLatest:%d:%d", remainLimit, coldCursor.UnixMilli())
			v, err, _ := fs.requestGroup.Do(sfKey, func() (interface{}, error) {
				return fs.repo.ListLatest(ctx, remainLimit, coldCursor)
			})
			if err == nil {
				coldVideos := v.([]*video.Video)
				baseVideos = append(baseVideos, coldVideos...)
			}
		}
	}
	var nextTime int64
	if len(baseVideos) > 0 {
		nextTime = baseVideos[len(baseVideos)-1].Createtime.UnixMilli()
	}
	hashmore := len(baseVideos) == limit
	feedVideos, err := fs.buildFeedVideos(ctx, baseVideos, viewerAccountID)
	if err != nil {
		return ListLatestResponse{}, err
	}
	return ListLatestResponse{
		VideoList: feedVideos,
		NextTime:  nextTime,
		HasMore:   hashmore,
	}, nil
}

// 按照点赞数查询视频
func (fs *FeedService) LikesCount(ctx context.Context, limit int, cursor *LikesCountCursor, viewerAccountID uint) (ListLikesCountResponse, error) {
	videos, err := fs.repo.ListLikesCount(ctx, limit, cursor)
	if err != nil {
		return ListLikesCountResponse{}, err
	}
	hasMore := len(videos) == limit
	feedItem, err := fs.buildFeedVideos(ctx, videos, viewerAccountID)
	if err != nil {
		return ListLikesCountResponse{}, err
	}
	resp := ListLikesCountResponse{
		VideoList: feedItem,
		HasMore:   hasMore,
	}
	if len(videos) > 0 {
		last := videos[len(videos)-1]
		nextLikesCountBefore := last.LikesCount
		nextIDBefore := last.ID
		resp.NextLikesCountBefore = &nextLikesCountBefore
		resp.NextIDBefore = &nextIDBefore
	}
	return resp, nil
}

// 按照关注列表查询视频
func (fs *FeedService) ListByFollowing(ctx context.Context, limit int, latestBefore time.Time, viewerAccountID uint) (ListByFollowingResponse, error) {
	doListByFollowingFromDB := func() (ListByFollowingResponse, error) {
		videos, err := fs.repo.ListByFollowing(ctx, limit, viewerAccountID, latestBefore)
		if err != nil {
			return ListByFollowingResponse{}, err
		}
		var nextTime int64
		if len(videos) > 0 {
			nextTime = videos[len(videos)-1].Createtime.Unix()
		} else {
			nextTime = 0
		}
		hasMore := len(videos) == limit
		feedVideos, err := fs.buildFeedVideos(ctx, videos, viewerAccountID)
		if err != nil {
			return ListByFollowingResponse{}, err
		}
		resp := ListByFollowingResponse{
			VideoList: feedVideos,
			NextTime:  nextTime,
			HasMore:   hasMore,
		}
		return resp, nil
	}
	var cacheKey string
	if viewerAccountID != 0 || fs.rediscache != nil {
		before := int64(0)
		if !latestBefore.IsZero() {
			before = latestBefore.Unix()
		}
		cacheKey = fmt.Sprintf("feed:listByFollowing:limit=%d:accountID=%d:before=%d", limit, viewerAccountID, before)
		cacheCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()

		b, err := fs.rediscache.GetBytes(cacheCtx, cacheKey)
		if err == nil {
			var cached ListByFollowingResponse
			if err = json.Unmarshal(b, &cached); err == nil {
				return cached, err
			}
		} else if rediscache.IsMiss(err) {
			//缓存未命中
			lockKey := "lock" + cacheKey
			token, locked, _ := fs.rediscache.Lock(cacheCtx, lockKey, 500*time.Millisecond)
			if locked {
				defer func() { _ = fs.rediscache.UnLock(cacheCtx, lockKey, token) }()
				if b, err = fs.rediscache.GetBytes(cacheCtx, cacheKey); err == nil {
					var cached ListByFollowingResponse
					if err = json.Unmarshal(b, &cached); err == nil {
						return cached, err
					}
				} else {
					//未命中缓存查数据库
					resp, err := doListByFollowingFromDB()
					if err != nil {
						return ListByFollowingResponse{}, err
					}
					if b, err := json.Marshal(&resp); err == nil {
						_ = fs.rediscache.SetBytes(cacheCtx, cacheKey, b, fs.cacheTTL)
					}
					return resp, nil
				}
			} else {
				for i := 0; i < 5; i++ {
					time.Sleep(20 * time.Millisecond)
					b, err = fs.rediscache.GetBytes(cacheCtx, cacheKey)
					if err == nil {
						var cached ListByFollowingResponse
						if err := json.Unmarshal(b, &cached); err == nil {
							return cached, nil
						}
					}
				}
			}
		}
	}
	resp, err := doListByFollowingFromDB()
	if err != nil {
		return ListByFollowingResponse{}, err
	}
	if cacheKey != "" {
		b, err := json.Marshal(&resp)
		if err == nil {
			_ = fs.rediscache.SetBytes(ctx, cacheKey, b, fs.cacheTTL)
		}
	}
	return resp, nil
}

// ListByPopularity 按热度获取视频列表（Redis热榜优先，DB降级游标分页）
func (fs *FeedService) ListByPopularity(ctx context.Context, limit int, reqAsOf int64, offset int, viewerAccountID uint, latestPopularity int64, latestBefore time.Time, latestIDBefore uint) (ListByPopularityResponse, error) {
	if fs.rediscache != nil {
		asOf := time.Now().UTC().Truncate(time.Minute)
		if reqAsOf > 0 {
			asOf = time.Unix(reqAsOf, 0).UTC().Truncate(time.Minute)
		}
		const win = 60
		keys := make([]string, 0, win)
		for i := 0; i < win; i++ {
			keys = append(keys, "hot:video:1m:"+asOf.Add(-time.Duration(i)*time.Minute).Format("200601021504"))

		}
		dest := "hot:video:1m" + asOf.Format("200601021504")
		opCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		exists, _ := fs.rediscache.Exists(opCtx, dest)
		if !exists {
			// ZUnionStore：合并多个 ZSET，按 SUM 聚合热度值（即视频近60分钟总热度）
			_ = fs.rediscache.ZUnionStore(opCtx, dest, keys, "SUM")
			_ = fs.rediscache.Expire(opCtx, dest, 2*time.Minute) // 留时间给翻页
		}
		// 从合并后的快照中分页获取视频 ID（ZRevRange 按热度倒序取）
		start := int64(offset)
		stop := start + int64(limit-1)
		members, err := fs.rediscache.ZRevRange(opCtx, dest, start, stop)
		if err == nil && len(members) == 0 {
			//无数据处理（offset>0 说明是翻页，直接返回无更多）
			if offset > 0 {
				return ListByPopularityResponse{
					VideoList:  []FeedVideoItem{},
					AsOf:       asOf.Unix(),
					NextOffset: offset,
					HasMore:    false,
				}, nil
			}
		}
		//有数据则解析 ID，查询视频详情并构造返回值

	}
	return ListByPopularityResponse{}, nil
}
