# Laretz agent guide

## Hard constraints

These are standing rules from the human reviewer. Violating any of them fails review.

### Workflow

- Implement one ticket at a time from the active tracker (`.scratch/<feature>/issues/`). After a ticket passes its gates: stop and present the diff for human review. Commit only after explicit approval.
- Gates per ticket, in order: `go build ./...`, `go vet ./...`, `golangci-lint run` (hard gate; `.golangci.yml` is law, never relax it), targeted `go test` per touched package, then full `go test ./...`. The full suite plus a code review run once after the last ticket.

### Layering (enforced by depguard in `.golangci.yml`)

- `internal/domain`: stdlib + shopspring/decimal only.
- `internal/app`: own tree + domain only; logr is the one allowed logging dependency. Ports are interfaces in `internal/app/interfaces`; use cases are `*UseCase` structs with `Execute(ctx, cmd)` in `internal/app/usecase`; sentinel errors in `internal/app` (e.g. `ErrEntityNotFound`). Data-access ports are repositories only (ADR-0005): one per domain concept, methods in domain language; no command/query payload layer between use case and port.
- `internal/impl`: app + domain + infra libraries; never server, lib, testing, shared, or cmd. Subpackages by concern: `database` (connection + goose migrations, embedded; sqlc `queries/` + `generated/`), `repository` (port implementations over sqlc; repositories own their transactions — multi-statement writes are atomic inside the impl), `plugin` (plugin host: plugin registry, host facade, `ExternalAssetProviderRegistry` + adapters); external plugin implementations live outside `internal/` (e.g. `plugin/coingecko`) behind the `pkg/plugin` API.
- `internal/shared`: cross-cutting support that may import infra libraries (logging, validator); consumable by server and impl.
- `internal/lib`: reusable utilities, stdlib only.
- `internal/app` and `internal/domain` are plain Go packages, never fx modules. fx lives only in `internal/server/module` (one `fx.Module` per concern); `internal/impl` exposes plain constructors wired by those modules.
- fx wiring names: `provide*` for funcs passed to fx.Provide, `register*` for fx.Invoke hooks; plain constructors (`New*`) may be handed to fx.Provide directly.

### Code style

- Least code wins: prefer a stdlib or chi built-in over custom code; generalize over duplicating.
- The binary entrypoint is `cmd/laretz/main.go`, a one-liner delegating to `cli.Run()`; all CLI parsing (urfave/cli v3) lives in `cmd/laretz/cli`; the app assembly (config, logger, fx start/stop) lives in `internal/server` (`app.go` + `module/`).
- Configuration is a yaml file in the data dir (`os.UserConfigDir()/laretz`), parsed and validated in `internal/server/config`; path override via `-c/--config` or `LARETZ_CONFIG`; `.env` is loaded by godotenv in `cmd/laretz/cli`.
- Middleware order in the router: origin check, then method spoof, then request log.
- Method spoofing is the `_method=DELETE` field in the POST body (ADR-0003); the spoof middleware lives in `internal/lib/spoof`; DELETE actions get a real chi `DELETE` route, the form POSTs to the same URL with the hidden field.
- Origin enforcement is a custom middleware; go-chi/cors only toggles response headers and never rejects a request (verified against its source). It lives in `internal/server/middleware`.
- Request logging is `shared/logging.NewLoggingMiddleware` (logr-backed), mounted in the router via `middleware.NewLoggingMiddleware`.
- Server-rendered pages go through `server/renderer`: a `*ResponseRenderer[PageData]` per page (Provider loads data via queries, Mutate adds errors); POST handlers re-render the page on validation errors with `views.Error{Message, Dialog}` instead of redirecting.
- Browser opening is opt-in via the `-o/--open` flag.
- Test helpers and mocks live in `internal/testing/` (`httpx` for request builders, `shared` for context-aware httptest requests, `mocks/` for generated mocks); never duplicate helpers across test packages.
- Never commit, copy, or log `.env` contents or API keys.

## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/` in this repo. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles, with label strings equal to their names. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.