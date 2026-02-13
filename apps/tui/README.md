# @libra-link/tui

Terminal-first Libra Link client built with Bubble Tea.

## Commands

- `bun run @libra-link/tui:dev` (via turbo) or `cd apps/tui && go run ./cmd/tui`
- `cd apps/tui && make build`
- `cd apps/tui && make test`
- `cd apps/tui && make sqlc-generate`
- `cd apps/tui && make api-generate`

## Environment

- `LIBRA_TUI_API_BASE_URL` (default `http://localhost:8080`)
- `LIBRA_TUI_DATA_DIR` (default `~/.local/share/libra-link-tui`)
- `LIBRA_TUI_HTTP_TIMEOUT_SECONDS` (default `15`)
- `LIBRA_TUI_SYNC_INTERVAL_SECONDS` (default `10`)
- `LIBRA_TUI_SYNC_BATCH_SIZE` (default `25`)
