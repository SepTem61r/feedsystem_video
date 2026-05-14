# feedsystem_video

视频 Feed 流系统，Go + Gin 后端，Vue 3 + TypeScript 前端。

## 项目结构

```
feedsystem_video/
├── backend/
│   ├── cmd/
│   │   ├── main.go                  # API 服务入口 (:8080)
│   │   └── worker/main.go           # Worker 消费端入口
│   ├── configs/
│   │   ├── config.yaml              # 本地开发配置 (localhost)
│   │   └── config.compose-local.yaml # Docker 部署配置
│   ├── internal/
│   │   ├── account/                 # 账号：注册/登录/改名/修改密码/登出
│   │   ├── auth/                    # JWT 生成与校验 (HS256, 24h)
│   │   ├── config/                  # YAML 配置加载 + 默认值兜底
│   │   ├── db/                      # MySQL 连接 + GORM AutoMigrate
│   │   ├── feed/                    # Feed 流：L1本地→L2Redis→L3MySQL 三级缓存
│   │   ├── http/                    # Gin 路由注册 
│   │   ├── middleware/
│   │   │   ├── jwt/                 # 硬校验 + 软校验中间件
│   │   │   ├── rabbitmq/            # MQ 生产者 (点赞/评论/热度/关注/时间线)
│   │   │   ├── ratelimit/           # 限流 (按IP/按账号)
│   │   │   └── redis/               # Redis 缓存 + ZSet + 分布式锁
│   │   ├── social/                  # 社交：关注/取关/粉丝列表/关注列表
│   │   └── video/                   # 视频：CRUD + 点赞 + 评论 + 热度缓存
│   └── worker/                      # MQ 消费者：点赞/评论/热度/关注/时间线
├── frontend/                        # 前端
├── Dockerfile                       # Go 后端多阶段构建
├── frontend/Dockerfile              # 前端构建 + nginx
├── frontend/nginx.conf              # nginx 反向代理配置
├── docker-compose.yaml              # 一键部署 (MySQL+Redis+RabbitMQ+Backend+Worker+Frontend)
├── go.mod / go.sum
└── .dockerignore
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端框架 | Go + Gin |
| ORM | GORM (AutoMigrate) |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 (三级缓存 L1本地 go-cache → L2 Redis → L3 MySQL) |
| 消息队列 | RabbitMQ (Topic 模式，异步写库 + 热度更新) |
| 认证 | JWT HS256 (24h 过期，Redis 存储支持吊销) |
| 限流 | Token Bucket (Redis INCR + TTL) |
| 前端 | Vue 3 + TypeScript + Vite + Pinia + Vue Router |
| 部署 | Docker Compose 六服务全容器化 |

## Feed 流三级缓存架构

```
请求 → L1 本地 go-cache (TTL 3s)
         ↓ miss
       L2 Redis (TTL 1h)
         ↓ miss
       L3 MySQL + singleflight 防击穿
```

- **冷热分离**：全局时间线 ZSet 维护热点窗口，冷数据降级查 MySQL
- **热度排行**：60 分钟滑动窗口 ZUnionStore 合并，2 分钟快照缓存
- **关注流**：Redis 缓存 + 分布式锁防击穿 + spin-lock 等待

## 异步写链路

```
API → RabbitMQ → Worker → MySQL
                     ↓
                Redis 热度 ZSet
```

- 点赞/评论/关注/取关先发 MQ 再返回 200，Worker 异步落库
- MQ 不可用时自动降级为同步写 MySQL
- Outbox 模式保证时间线事件不丢失

## 前端页面

| 路由 | 页面 | 说明 |
|---|---|---|
| `/` | SwipeFeed | 抖音式全屏上下滑，底部 Tab：最新/最热/关注/点赞/我 |
| `/explore` | FeedPage | 多 Tab 网格 Feed |
| `/video/:id` | 视频详情 | 播放 + 点赞 + 评论 + 删除 |
| `/profile/:id` | 个人主页 | 作品/点赞/关注/粉丝 + 修改用户名/密码 |
| `/upload` | 发布视频 | 视频 + 封面上传 |
| `/login` | 登录 | |
| `/register` | 注册 | |

## API 接口 (32个)

| 模块 | 端点 | 方法 | 认证 |
|---|---|---|---|
| Account | `/account/register` `/login` `/changePassword` `/findByID` `/findByUsername` | POST | 公开 |
| Account | `/account/logout` `/rename` | POST | JWT |
| Video | `/video/listByButhorID` `/getDetail` | POST | 公开 |
| Video | `/video/publish` `/uploadVideo` `/uploadCover` `/delete` | POST | JWT |
| Like | `/like/like` `/unlike` `/isLiked` `/listMyLikedVideos` | POST | JWT |
| Comment | `/comment/getAll` | POST | 公开 |
| Comment | `/comment/publish` `/delete` | POST | JWT |
| Social | `/social/follow` `/unfollow` `/getallfollowers` `/getallvloggers` | POST | JWT |
| Feed | `/feed/listLatest` `/listLikesCount` `/listByPopularity` | POST | 软 JWT |
| Feed | `/feed/listByFollowing` | POST | JWT |
