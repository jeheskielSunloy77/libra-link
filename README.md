# libra-link

A monorepo for a Go API and Go TUI client with shared TypeScript packages, managed with Turborepo and Bun workspaces. scaffolded with **libra-link** visit the
[repository](https://github.com/jeheskielSunloy77/libra-link) for more details.

## Repository layout

```
libra-link/
├── apps/api             # Go API (Fiber)
├── apps/tui             # Go TUI client (Bubble Tea + SQLite/sqlc)

├── packages/zod         # Shared Zod schemas
├── packages/openapi     # OpenAPI generation
├── packages/emails      # React Email for email templates generation
└── packages/*           # Other shared packages
```

## Prerequisites

- Go 1.24+
- Bun 1.2.13 (Node 22+)
- PostgreSQL 16+
- Redis 8+

## Quick start

```bash
bun install                          # Install dependencies for all apps and packages
cp apps/api/.env.example apps/api/.env      # Set up API env

bun run api:migrate:up   # Run DB migrations

# Start all apps
bun dev

# Run only the TUI app
bun run tui:run
```

Or you can use docker compose for local development:

```bash
docker compose up --build
```

## Common commands

```bash
# Monorepo (from root)
bun dev         # Start dev servers for all apps
bun dev:all    # Start dev servers for all apps and packages
bun run test       # Run tests for all apps and packages
bun run build
bun run lint
bun run typecheck

# API helpers (see apps/api/Makefile for migrate targets)
bun run api:dev
bun run api:test
cd apps/api && make migrate-new NAME=add_table
cd apps/api && make migrate-up
cd apps/api && make migrate-down

# TUI helpers
bun run tui:run
cd apps/tui && make build
cd apps/tui && make test
cd apps/tui && make sqlc-generate
cd apps/tui && make api-generate

# Contracts and emails
bun run openapi:generate    # Generate OpenAPI spec file from contracts
bun run emails:generate     # Generate email HTML templates


```

## API (apps/api)

### Technologies

- Fiber web framework
- GORM ORM with PostgreSQL
- Asynq for background jobs with Redis
- Zap + Zerolog for logging
- Testcontainers for integration tests
- OpenAPI documentations UI
- SMTP email handling
- Redis caching layer

### Architecture & Conventions

- Clean layers: handlers -> services -> repositories -> models.
- Repositories are data access only; services implement business rules and validations.
- Use `ResourceRepository` / `ResourceService` / `ResourceHandler` for standard CRUD models.
- Entry points: `apps/api/cmd/api/main.go` (server) and `apps/api/cmd/seed/main.go` (seeder).
- Routes in `apps/api/internal/router/routes.go`; middleware order in `apps/api/internal/router/router.go`.
- Prefer `handler.Handle` / `handler.HandleNoContent` / `handler.HandleFile` for new endpoints.
- Request DTOs implement `validation.Validatable`; use `validation.BindAndValidate` or the handler wrappers.
- Use `utils.ParseUUIDParam` for `:id` params.
- Services return `errs.ErrorResponse`; wrap DB errors with `sqlerr.HandleError`. Handlers return errors and let `GlobalErrorHandler` format responses.
- Request IDs are set in middleware and injected into logs; use `middleware.GetLogger` in handlers.
- Context timeouts should use `server.Config.Server.ReadTimeout` / `WriteTimeout`.
- Auth uses short-lived JWT access tokens and long-lived refresh tokens. `middleware.Auth.RequireAuth` sets `user_id` in Fiber locals; sessions live in `auth_sessions`. Cookie config is under `AuthConfig`.
- Auth routes: `/api/v1/auth/register`, `/login`, `/google`, `/google/device/start`, `/google/device/poll`, `/verify-email`, `/refresh`, `/me`, `/resend-verification`, `/logout`, `/logout-all`.
- Background jobs use Asynq (`apps/api/internal/lib/job`). Define new task payloads in `email_tasks.go`, register them in `JobService.Start`, and wire handlers in `handlers.go`.
- Email templates live in `apps/api/templates/emails` and are generated from `packages/emails`.
- OpenAPI docs are written to `apps/api/static/openapi.json` and served at `/api/docs`. Update `packages/zod` and `packages/openapi/src/contracts` when endpoints change.
- Caching layer with Redis in `apps/api/internal/lib/cache`.

## TUI (apps/tui)

### Technologies

- Bubble Tea + Bubbles + Lipgloss
- SQLite + sqlc for local cache and sync outbox
- OpenAPI-generated Go client from `apps/api/static/openapi.json`

### Architecture & Conventions

- Entry point: `apps/tui/cmd/tui/main.go`.
- UI/state orchestration: `apps/tui/internal/app`.
- Local persistence and queries: `apps/tui/internal/storage/sqlite`.
- Sync retry worker: `apps/tui/internal/sync`.
- Reader behavior and format adapters: `apps/tui/internal/reader`.
- TUI auth uses browser-assisted Google device flow via API endpoints.

### Local Environment Variables

- `LIBRA_TUI_API_BASE_URL` (default `http://localhost:8080`)
- `LIBRA_TUI_DATA_DIR` (default `~/.local/share/libra-link-tui`)
- `LIBRA_TUI_HTTP_TIMEOUT_SECONDS` (default `15`)
- `LIBRA_TUI_SYNC_INTERVAL_SECONDS` (default `10`)
- `LIBRA_TUI_SYNC_BATCH_SIZE` (default `25`)



## Packages (packages/\*)

- `@libra-link/zod` (`packages/zod`): source of truth for API request/response schemas (exported from `packages/zod/src/index.ts`).
- `@libra-link/openapi` (`packages/openapi`): builds the OpenAPI spec from Zod + ts-rest contracts in `packages/openapi/src/contracts`. Regenerate with `bun run openapi:generate`.

- `@libra-link/emails` (`packages/emails`): React Email templates in `packages/emails/src/templates`. Export HTML to `apps/api/templates/emails` via `bun run emails:generate`.

## Testing

- Services: unit tests only, mock repositories.
- Repositories: integration tests with real PostgreSQL (Testcontainers), no SQL mocking.
- Handlers: thin HTTP tests only, mock services.
- Tests live next to code (`foo.go` -> `foo_test.go` / `foo_integration_test.go`).
- Use helpers in `apps/api/internal/testing` (`SetupTestDB`, `WithRollbackTransaction`).

## DevOps

- This project is designed to be containerized. it is already dockerized with Dockerfiles in `apps/api/Dockerfile`.
- Use docker compose file on `docker-compose.yml` for local development with containers.
- CI/CD is set up with GitHub Actions in `.github/workflows/ci.yml`.
