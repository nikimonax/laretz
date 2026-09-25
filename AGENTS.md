# Laretz agent guide

## Hard constraints

These are standing rules from the human reviewer. Violating any of them fails review.

### Workflow

- Implement one ticket at a time from the active tracker (`.scratch/<feature>/issues/`). After a ticket passes its gates: stop and present the diff for human review. Commit only after explicit approval.
- Gates per ticket, in order: `go build ./...`, `go vet ./...`, `golangci-lint run` (hard gate; `.golangci.yml` is law, never relax it), targeted `go test` per touched package, then full `go test ./...`. The full suite plus a code review run once after the last ticket.
- Test-first at the pre-agreed seams: domain services, CQS commands/queries against sqlite, middleware, external HTTP adapters against fake servers.

### Layering (enforced by depguard in `.golangci.yml`)

- `internal/domain`: stdlib + shopspring/decimal only.
- `internal/app`: own tree + domain only; logr is the one allowed logging dependency.
- `internal/impl`: app + domain + infra libraries; never server, lib, testing, or cmd.
- `internal/lib`: reusable utilities, stdlib only.
- `internal/app` and `internal/domain` are plain Go packages, never fx modules. fx lives in `internal/impl` (infrastructure modules) and `internal/server` (composition root).
- fx wiring names: `provide*` for funcs passed to fx.Provide, `register*` for fx.Invoke hooks; plain constructors (`New*`) may be handed to fx.Provide directly.

### Code style

- Least code wins: prefer a stdlib or chi built-in over custom code; generalize over duplicating.
- Root `main.go` stays a one-liner delegating to `cmd.Execute()`; all CLI parsing lives in the `cmd` package; the app entrypoint and assembly (logger, fx start/stop) live in `internal/server`.
- Middleware order in the router: request log, then method spoof, then origin check.
- Method spoofing is the `_method=DELETE` field in the POST body (ADR-0003); the spoof middleware lives in `internal/lib/spoof`.
- Origin enforcement is a custom middleware; go-chi/cors only toggles response headers and never rejects a request (verified against its source). It lives in `internal/server/middleware`.
- Request logging goes through chi `middleware.RequestLogger` with a logr formatter.
- Browser opening is opt-in via the `-o/--open` flag.
- Test helpers and mocks live in `internal/testing/` (`httpx` for request builders, `mocks/` for generated mocks); never duplicate helpers across test packages.
- Never commit, copy, or log `.env` contents or API keys.

## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/` in this repo. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles, with label strings equal to their names. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.