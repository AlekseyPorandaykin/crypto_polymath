# Cursor context — Crypto Polymath

Read this file at the start of a session or before non-trivial work. It orients you;
canonical rules live in [`.ai/`](../.ai/README.md) and must not be duplicated here.

## What this is

Educational non-commercial server for crypto technical analysis on public market data.
One Go binary (`main.go` → `cmd/`), three processes:

| Process | Command | Role |
| --- | --- | --- |
| Loader | `daemon loader` | Fetch prices/candles/metadata from exchanges → DB + events |
| Calculator | `daemon calculator` | Consume events → indicators and analytics |
| API | `api external-v1` | REST, web UI, OpenAPI docs |

Not a trading bot: no orders, no exchange API keys from users. API serves what the
loader already stored.

Module: `github.com/AlekseyPorandaykin/crypto_polymath`, Go 1.24, vendored deps.

## Where code lives

| Layer | Path | Rule |
| --- | --- | --- |
| Domain types | `domain/` | No project imports |
| Business logic | `core/` | No HTTP/SQL/exchange clients |
| Implementations | `internal/infrastructure/` | Repos, adapters, queues, cache |
| Entry points | `internal/ui/` | Daemons, REST, gRPC, web — thin |
| DI wiring | `cmd/container/` | `go.uber.org/dig` |
| REST contract | `api/rest/v1/openapi.yaml` | Source of truth; regenerate, never hand-edit `*.gen.go` |
| Shared helpers | `pkg/` | No domain knowledge |

`core/` packages: `price`, `candlestick`, `candle_indicator`, `indicator`, `analysis`,
`exchange`, `trading` (pure position maths — outside the market cascade).

## Mandatory reading (in order)

1. [`.ai/project.md`](../.ai/project.md) — product boundaries
2. [`.ai/structure.md`](../.ai/structure.md) — map and where new code goes
3. [`.ai/conventions.md`](../.ai/conventions.md) — language, tests, contract-first, frontend
4. [`.ai/journal.md`](../.ai/journal.md) — last few entries (current state)

Architecture depth: [`docs/overview.md`](../docs/overview.md).
Core dependency DAG / calculation stages: [`docs/core-pipeline.md`](../docs/core-pipeline.md).

## Hard rules (do not skip)

- After any code change: `go test -count=1 ./tests/bdd/...`. On failure, report the
  conflict; do not quietly rewrite expectations.
- REST: edit `openapi.yaml` → regenerate `internal/ui/api/v1/spec/*.gen.go` → implement
  handler. Never edit generated files by hand.
- Language: Go comments/tests/BDD in **Russian**; OpenAPI + `internal/ui/web/static/**` in
  **English** (enforced by tests).
- Comments explain **why**, never what, never the change you just made.
- Do not commit unless the user asked.
- When a task is done, append an entry to `.ai/journal.md` (template at top of that file).

## Cursor-specific rules

Always-applied and scoped rules under [`.cursor/rules/`](rules/):

- `ai-context.mdc` — points here and to `.ai/`
- `testing-workflow.mdc` — BDD after every change
- `api-contract.mdc` — when editing OpenAPI / `internal/ui/api/v1/**`
- `web-static.mdc` — when editing `internal/ui/web/**`

Shared entry points for other assistants: `AGENTS.md`, `CLAUDE.md` → same `.ai/` folder.

## Quick placement guide

| Task | Put it here |
| --- | --- |
| New indicator / calculation | `core/<area>/` + unit tests + BDD in `tests/bdd/features/` |
| New REST endpoint | `api/rest/v1/openapi.yaml` → regenerate → `internal/ui/api/v1/impl/` |
| New exchange adapter | `internal/infrastructure/adapters/` + wire in `cmd/container/` |
| New daemon job | `internal/ui/daemon/` |
| New page / asset | `internal/ui/web/static/` + route in `pages.go` |
| New domain type | `domain/` only |

## Pitfalls worth knowing

- `go test ./internal/ui/...` can hang (daemon tests hit the network) — test narrower packages.
- `go vet ./...` may report pre-existing unreachable code in
  `internal/ui/daemon/candlesticks_consumer/app.go`.
- Landing page may carry only `static/js/contact.js`; keep it static for crawlers.
- Tools page: Vue/PrimeVue versions must match in both pin sites; use `render()`/`h()`, not
  string templates.
- Token (`X-Token`) is not validated yet — keep that fact out of the OpenAPI wording.
