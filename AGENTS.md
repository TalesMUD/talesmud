# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` houses Go entrypoints (server and dialog sandbox).
- `pkg/` contains core Go packages (DB, services, scripting, embedded web UI).
- `public/app/` is the Svelte/Vite admin frontend; `public/mud-client/` is the Svelte/Rollup game client.
- `api/` stores HTTP request examples and import/export fixtures.
- `data/`, `designs/`, and `game-design/` hold content/design assets; `bin/` is the build output location.

## Build, Test, and Development Commands
- `make build`: builds both frontends, copies dist assets into `pkg/`, and compiles `bin/tales`.
- `make build-frontend` / `make build-mud-client` / `make build-backend`: build each target individually.
- `make run-server`: run the Go server locally.
- `make run-frontend`: start the admin UI (`public/app/`) via Vite.
- `make run-mud-client`: start the game client in watch mode.
- `make run`: run server + both frontends concurrently.

## Architecture & Configuration
- The server is split between a Gin HTTP API (`pkg/server/`) and the WebSocket MUD server/game loop (`pkg/mudserver/`).
- Business logic lives in `pkg/service/` with repositories in `pkg/repository/` for MongoDB/SQLite.
- Set `DB_DRIVER=mongo|sqlite` and `SQLITE_PATH` (SQLite). Auth uses Auth0 JWTs.
- Scripts run via Otto JS in `pkg/scripts/`; see `game-design/SCRIPTING.md` for hooks.

## Coding Style & Naming Conventions
- Go code follows standard `gofmt` formatting; keep packages lowercase and file names short and descriptive.
- Keep exported Go identifiers in `CamelCase`, unexported in `camelCase`.
- Frontend code follows existing Svelte/Vite or Svelte/Rollup conventions; prefer `camelCase` for JS variables and `kebab-case` for CSS classes.

## Testing Guidelines
- Go tests live alongside packages as `*_test.go` (e.g., `pkg/db/db_test.go`).
- Run tests with `go test ./...`; for focused runs, use `go test ./pkg/db`.
- The repo uses `testify` assertions; keep tests deterministic.

## Commit & Pull Request Guidelines
- Recent commits use short, imperative summaries (e.g., “Refactor database handling…”).
- Avoid “WIP” in mainline history; squash before merge if needed.
- PRs should include: summary, how to test, and screenshots/gifs for UI changes in `public/`.

## Configuration & Data Notes
- Local data files (e.g., `talesmud.db`) live at the repo root; avoid committing generated data.
- When adding new content assets, place them under `data/` or `designs/` and document paths in PRs.

## Documentation & Design References
- `PROJECT.md` and `ARCHITECTURE.md` summarize system goals, routes, and major components.
- The `game-design/` folder tracks the MVP backlog, scripting system, and world map implementation details.
