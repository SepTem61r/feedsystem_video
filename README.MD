# feedsystem_video

视频流推荐系统后端，Go + Gin 框架。

## 项目结构

```text
feedsystem_video/
├── backend/                    # Go 后端
│   ├── go.mod / go.sum         # Go 模块依赖
│   ├── cmd/
│   │   └── main.go             # API 服务入口（端口 8080）
│   ├── configs/                # YAML 配置文件
│   │   └── config.compose-local.yaml
│   └── internal/               # 核心业务代码
│       ├── account/            # 账号模块（注册、登录、改名、修改密码）
│       ├── auth/               # JWT 令牌生成与校验
│       ├── config/             # 配置加载（YAML + 默认值兜底）
│       ├── db/                 # 数据库连接 + GORM AutoMigrate
│       ├── http/               # 路由注册（Gin router，25 条路由）
│       ├── middleware/
│       │   ├── jwt/            # JWT 认证中间件（硬校验 / 软校验）
│       │   ├── rabbitmq/       # MQ 生产者（点赞/评论/热度/关注）
│       │   ├── ratelimit/      # 限流中间件（按 IP / 按账号）
│       │   └── redis/          # Redis 缓存 + ZSet + 分布式锁
│       ├── social/             # 社交模块（关注/取关/粉丝/关注列表）
│       └── video/              # 视频模块（CRUD、点赞、评论、热度缓存）
├── frontend/                   # 前端（Vite + React/Vue，待开发）
├── data/                       # 初始化数据 / SQL
├── test/                       # 测试文件
└── docker-compose.yml          # 一键启动（待开发）
```

## 架构

| 层 | 技术 | 说明 |
|---|------|------|
| HTTP 框架 | Gin | RESTful API，路由按模块分组 |
| ORM | GORM | AutoMigrate 自动建表 |
| 数据库 | MySQL 8.0 | 持久化存储 |
| 缓存 | Redis 7 | 登录状态、热度排行(ZSet)、限流计数 |
| 消息队列 | RabbitMQ | API 进程写 MQ，Worker 进程异步消费落库 |
| 认证 | JWT (HS256) | 硬校验（必须登录）与软校验（可选登录）两种中间件 |
| 限流 | Token Bucket (Redis) | 登录/注册按 IP，写操作按账号 |

## 当前进度

已完成：account、video（含 like/comment）、social 全部 CRUD + 路由注册 + AutoMigrate + JWT + Redis + RabbitMQ。

待开发：feed 流、worker 消费端、pprof、Dockerfile、前端。

## 部署方式

- Docker Compose 全容器化：`docker compose up -d --build`，前端 5173、API 8080、RabbitMQ 管理台 15672
- 本地混合开发：compose 只启 MySQL/Redis/RabbitMQ，Go 和前端各自本地启动
- 纯本地开发：`cd backend && go run ./cmd/`，依赖 `DefaultLocalConfig()` 兜底配置

## 核心数据流

API 进程处理请求后，将写操作（点赞、评论、关注、视频发布）的消息投递到 RabbitMQ，Worker 进程异步消费这些消息，执行数据库写入和 Redis 热度缓存更新，实现读写分离和削峰。
