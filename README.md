# RealWorld API 服务

这是一个基于 Go、Gin、MySQL、Redis 实现的 RealWorld / Conduit 风格后端项目，并内置了一个轻量级前端页面，方便本地联调和日常使用。

项目当前重点包括：

- 清晰的请求链路：`middleware -> handler -> biz -> store -> MySQL/Redis`
- 基于 `/api` 的 RealWorld 风格接口
- 公共读取接口支持可选登录态
- 优先支持 Docker 本地启动
- Redis + 本地缓存
- 健康检查和指标接口

## 功能概览

- 用户注册、登录、获取当前用户、更新用户
- JWT access token + refresh token 机制
- 用户资料查询、关注、取消关注
- 文章创建、列表、详情、更新、删除
- 按作者、标签、收藏用户筛选文章
- 已登录用户的个人 Feed
- 评论创建、查询、删除
- 标签列表
- 健康检查：`/healthz`、`/readyz`
- 指标接口：`/metrics`、`/metrics/concurrency`、`/metrics/cache`
- 内嵌浏览器 UI：`/ui/`

## 技术栈

- Go
- Gin
- GORM
- MySQL 8
- Redis 7
- Docker Compose

## 目录结构

```text
api/            API 契约文档与 Postman 集合
apiserver/      HTTP 服务、handler、biz、store、middleware、UI
common/         通用数据库辅助代码
config/         配置加载与环境变量覆盖
docs/           运行与架构说明
db.sql          数据库初始化脚本
compose.yaml    本地全栈 Docker 编排
Dockerfile      多阶段构建镜像
```

## 快速开始

### 方式一：Docker

先复制一份环境文件并替换其中的占位密码和 JWT 密钥：

```bash
cp .env.example .env
```

`.env.example` 里的值只是安全占位符，不应该直接用于共享环境、云环境或公网机器。

```bash
docker compose up --build -d
```

默认端口：

- API：`http://localhost:18080`
- MySQL：`127.0.0.1:13306`
- Redis：`127.0.0.1:16379`

常用检查：

```bash
curl http://localhost:18080/healthz
curl http://localhost:18080/readyz
curl http://localhost:18080/api/articles
curl http://localhost:18080/api/tags
curl http://localhost:18080/metrics/cache
```

### 方式二：本地运行

先启动 MySQL 和 Redis，再执行：

```bash
go run ./apiserver
```

如果使用仓库里的 `config.yaml`，请先把其中的占位密码和 `jwt.secret` 改成你自己的本地值，或者直接通过环境变量覆盖。

示例环境变量：

```powershell
$env:CONFIG_PATH = "./config.yaml"
$env:MYSQL_ADDR = "127.0.0.1:3306"
$env:REDIS_ADDR = "127.0.0.1:6379"
$env:JWT_SECRET = "change-this-local-dev-jwt-secret"
```

## 内嵌前端

项目内置了一个简易前端页面，便于手工验证接口和日常操作。

- 首页入口：`http://localhost:18080/`
- UI 直达：`http://localhost:18080/ui/`

目前支持：

- 注册 / 登录
- 获取当前用户
- 刷新 token
- 文章创建 / 编辑 / 删除
- 文章筛选
- 收藏文章
- 评论创建 / 删除

## API 说明

- API 前缀：`/api`
- 鉴权头格式：`Authorization: Token <token>`
- 公共读取接口支持可选登录态
- 标准评论删除路由：

```text
DELETE /api/articles/:slug/comments/:id
```

- 为兼容旧调用，仍保留：

```text
DELETE /api/comments/:id
```

## 限流

当前 API 限流与仓库契约保持一致：

- 每分钟 60 次请求
- 每小时 1000 次请求

已登录请求按用户或 token 计数，匿名请求按客户端 IP 计数。

## 配置

配置先从 `config.yaml` 加载，再由环境变量覆盖。

注意：

- 仓库里的 `config.yaml`、`apiserver/config.yaml`、`.env.example` 都只保留了占位符，不能视为生产可用配置
- 不要把真实数据库密码、JWT secret 或其他密钥提交到 Git
- 如果真实密钥已经提交过，当前修改只能阻止继续泄露，不能消除历史中的旧值；处理步骤见 `docs/secret-cleanup.md`

常用环境变量：

- `SERVER_PORT`
- `SERVER_RATE_LIMIT_PER_MINUTE`
- `SERVER_RATE_LIMIT_PER_HOUR`
- `MYSQL_ADDR`
- `MYSQL_USERNAME`
- `MYSQL_PASSWORD`
- `MYSQL_DATABASE`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `JWT_SECRET`

## 测试

默认测试：

```bash
go test ./...
```

集成测试需要服务已启动：

```bash
go test -tags=integration ./apiserver/test
```

此外还补了接口契约对齐测试，覆盖：

- 用户响应体字段
- 更新用户时的冲突处理
- 文章分页参数
- 删除接口状态码
- 限流行为

## 契约对齐说明

当前实现已经和仓库中的 API 契约对齐了这些关键点：

- `PUT /api/user` 在用户名或邮箱冲突时返回 `400`
- `GET /api/articles` 支持 `page` + `limit`
- 删除文章和删除评论成功时返回 `200 OK`
- 用户认证相关响应体只返回契约定义字段
- refresh token 通过 `X-Refresh-Token` 响应头暴露，兼容当前 UI 使用

## 开发说明

- `docs/runtime.md` 记录了运行链路和请求链路
- `docs/secret-cleanup.md` 记录了敏感信息轮换与 Git 历史清理方案
- `api/api.md` 是接口契约
- `api/Conduit.postman_collection.json` 可用于手工接口验证
