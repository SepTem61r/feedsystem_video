# feedsystem_video

视频 Feed 流系统，Go + Gin 后端，Vue 3 + TypeScript 前端。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端框架 | Go + Gin |
| ORM | GORM  |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 (三级缓存 L1本地 go-cache → L2 Redis → L3 MySQL) |
| 消息队列 | RabbitMQ (Topic 模式，异步写库 + 热度更新) |
| 认证 | JWT HS256 (24h 过期，Redis 存储支持吊销) |
| 限流 | Token Bucket (Redis INCR + TTL) |
| 前端 | Vue 3 + TypeScript + Vite + Pinia + Vue Router |
| 部署 | Docker Compose 六服务全容器化 |

## 快速开始

### Docker 一键部署

```bash
docker-compose up -d
```

访问 `http://localhost:3000`。

服务列表：

| 服务 | 端口 |
|---|---|
| 前端 (nginx) | 3000 |
| 后端 API | 8080 |
| Worker 消费端 | — |
| MySQL | 3306 |
| Redis | 6379 |
| RabbitMQ 管理面板 | 15672 |

### 本地混合开发

```bash
# 1. 启动基础设施
docker-compose up -d mysql redis rabbitmq

# 2. 启动后端
go run ./backend/cmd/main.go

# 3. 另开终端，启动 Worker
go run ./backend/cmd/worker/main.go

# 4. 启动前端
cd frontend && npm run dev
```

前端 `http://localhost:5173`，Vite 代理自动转发 API 到 `:8080`。

配置文件在 `backend/configs/config.yaml`（本地 all-localhost）和 `backend/configs/config.compose-local.yaml`（Docker 服务名）。