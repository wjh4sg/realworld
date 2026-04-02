# Runtime Guide

## Architecture

Startup flow:

1. `apiserver/main.go` loads config from `config.yaml`.
2. Environment variables override YAML values for container and CI use.
3. MySQL and Redis are initialized.
4. Redis + Ristretto cache manager is created and warmed.
5. `store` is assembled, then `biz`, then `handler`.
6. Gin routes are registered under `/api`.
7. Health and metrics endpoints are exposed at the root path.

Request flow:

1. `Cors`
2. `RateLimiter`
3. `Auth` or `OptionalAuth`
4. handler request binding and validation
5. biz rules
6. store queries and cache reads/writes
7. MySQL / Redis

## API Notes

- Public read endpoints now use optional auth, so the same endpoint can return `favorited` / `following` when a valid `Authorization: Token ...` header is present.
- Article responses now come from persisted data only. Placeholder article bodies, authors, and injected defaults were removed.
- Refresh now accepts refresh tokens only.
- Canonical comment deletion route is `DELETE /api/articles/:slug/comments/:id`.
- Legacy compatibility route `DELETE /api/comments/:id` is still available.
- Health probes:
  - `GET /healthz`: process is alive
  - `GET /readyz`: MySQL and Redis are reachable

## Local Run

Start dependencies yourself, then run:

```bash
go run ./apiserver
```

Before local startup, replace the placeholder MySQL password and JWT secret in `config.yaml`, or override them with environment variables. The tracked config files are templates, not trusted secrets.

Useful environment overrides:

```powershell
$env:CONFIG_PATH = "./config.yaml"
$env:MYSQL_ADDR = "127.0.0.1:3306"
$env:REDIS_ADDR = "127.0.0.1:6379"
$env:JWT_SECRET = "change-this-local-dev-jwt-secret"
```

## Docker Run

Copy `.env.example` to `.env`, replace the placeholder passwords and JWT secret, then run:

```bash
docker compose up --build
```

The values committed in `.env.example` and `compose.yaml` are safe placeholders for local setup only. Treat them as examples, not as deployed credentials.

Default host ports:

- app: `http://localhost:18080`
- mysql: `127.0.0.1:13306`
- redis: `127.0.0.1:16379`

Useful checks:

```bash
curl http://localhost:18080/healthz
curl http://localhost:18080/readyz
curl http://localhost:18080/api/articles
curl http://localhost:18080/api/tags
curl http://localhost:18080/metrics/cache
```

Simple browser UI:

- `http://localhost:18080/`
- `http://localhost:18080/ui/`

## Test Layers

- Default: `go test ./...`
  - fast unit and package tests only
- Integration: `go test -tags=integration ./...`
  - requires a running service stack
- Performance: `go test -tags=perf ./apiserver/test`
  - manual execution only

Suggested Docker validation flow:

1. `docker compose up --build -d`
2. wait for `http://localhost:18080/readyz`
3. run key API smoke checks
4. optionally run Postman / Newman assets in `api/`
5. optionally run perf-tagged tests
