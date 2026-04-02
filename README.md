# RealWorld API Server

A production-style RealWorld backend implemented with Go, Gin, MySQL, Redis, and a lightweight embedded web UI.

This project focuses on:

- clean request flow: `middleware -> handler -> biz -> store -> MySQL/Redis`
- RealWorld / Conduit-style API behavior under `/api`
- optional auth on public read endpoints
- Docker-first local startup
- Redis + in-process cache support
- health and metrics endpoints for validation and debugging

## Features

- User registration, login, current user, profile update
- JWT access token + refresh token flow
- Profile lookup, follow, unfollow
- Article create, list, detail, update, delete
- Tag filtering, author filtering, favorited filtering
- Feed endpoint for authenticated users
- Comment create, list, delete
- Tags endpoint
- Health endpoints: `/healthz`, `/readyz`
- Metrics endpoints: `/metrics`, `/metrics/concurrency`, `/metrics/cache`
- Embedded browser UI at `/ui/`

## Tech Stack

- Go
- Gin
- GORM
- MySQL 8
- Redis 7
- Docker Compose

## Project Structure

```text
api/            API contract and Postman collection
apiserver/      HTTP server, handlers, biz, store, middleware, UI
common/         shared database helpers
config/         config loading and env override logic
docs/           runtime and architecture notes
db.sql          database schema and init script
compose.yaml    local full-stack Docker setup
Dockerfile      multi-stage application image
```

## Quick Start

### Option 1: Docker

```bash
docker compose up --build -d
```

Default services:

- API: `http://localhost:18080`
- MySQL: `127.0.0.1:13306`
- Redis: `127.0.0.1:16379`

Useful checks:

```bash
curl http://localhost:18080/healthz
curl http://localhost:18080/readyz
curl http://localhost:18080/api/articles
curl http://localhost:18080/api/tags
curl http://localhost:18080/metrics/cache
```

### Option 2: Local run

Start MySQL and Redis first, then run:

```bash
go run ./apiserver
```

Example environment overrides:

```powershell
$env:CONFIG_PATH = "./config.yaml"
$env:MYSQL_ADDR = "127.0.0.1:3306"
$env:REDIS_ADDR = "127.0.0.1:6379"
$env:JWT_SECRET = "replace-me"
```

## Embedded UI

The project includes a simple embedded frontend for manual testing and day-to-day usage.

- Root entry: `http://localhost:18080/`
- Direct UI path: `http://localhost:18080/ui/`

The UI supports:

- register / login
- current user lookup
- token refresh
- article create / edit / delete
- article filtering
- favorites
- comment create / delete

## API Notes

- API base path: `/api`
- Auth header format: `Authorization: Token <token>`
- Public read routes support optional auth
- Canonical comment delete route:

```text
DELETE /api/articles/:slug/comments/:id
```

- Compatibility route kept for older clients:

```text
DELETE /api/comments/:id
```

## Rate Limiting

API requests are limited with fixed windows that match the repository contract:

- 60 requests per minute
- 1000 requests per hour

Authenticated requests are keyed per user/token, while anonymous requests fall back to client IP.

## Configuration

Configuration is loaded from `config.yaml`, then overridden by environment variables.

Important variables:

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

## Tests

Default test run:

```bash
go test ./...
```

Integration tests require a running service stack:

```bash
go test -tags=integration ./apiserver/test
```

There are also dedicated handler and middleware contract-alignment tests covering:

- user response body shape
- update-user conflict handling
- article pagination parameters
- delete status codes
- rate limiting behavior

## Contract Alignment

The implementation has been aligned with the repository API contract for these key items:

- `PUT /api/user` returns `400` on duplicate username/email
- `GET /api/articles` supports `page` + `limit`
- successful article/comment deletes return `200 OK`
- user auth responses return only contract fields in JSON body
- refresh token is exposed through the `X-Refresh-Token` response header for UI compatibility

## Development Notes

- `docs/runtime.md` contains the request flow and runtime notes
- `api/api.md` contains the API contract
- `api/Conduit.postman_collection.json` contains manual API assets

