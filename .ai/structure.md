# Repository map

Module: `github.com/AlekseyPorandaykin/crypto_polymath`, Go 1.24. Dependencies are
vendored (`vendor/`), so builds do not reach the network.

```
crypto_polymath/
├── main.go                 # entry point, hands over to cmd/
├── cmd/                    # cobra commands: daemon loader, daemon calculator, api external-v1, migrate
│   └── container/          # DI container (dig): connections, repositories, services, applications
├── domain/                 # pure domain types and constants: Candlestick, Indicator, Unit (m/H/D/W/M)
├── core/                   # business logic, no HTTP and no SQL
│   ├── price/  candlestick/  indicator/  analysis/  candle_indicator/  exchange/
│   └── trading/            # position maths: entry price, liquidation, PnL, risk
├── internal/
│   ├── config/             # configuration loading (viper)
│   ├── event/              # event listeners, publishing into queues
│   ├── infrastructure/     # implementations: postgresql/, sqlite/, rabbitmq/, memory/, adapters/, observability/
│   └── ui/                 # everything that faces the outside world
│       ├── api/v1/spec/    # GENERATED from the OpenAPI contract — never edit by hand
│       ├── api/v1/impl/    # REST handlers implementing spec.ServerInterface
│       ├── api/grpc/       # gRPC service (generated action/ package plus handlers)
│       ├── daemon/         # long-running processes: loader, calculator, candlesticks_consumer
│       └── web/            # embedded web interface: pages.go plus static/
├── api/
│   ├── rest/v1/            # openapi.yaml — the source of truth for the REST API
│   ├── rest/client/        # generated client models for external consumers
│   ├── grpc/v1/            # .proto definitions
│   └── queue/              # queue message schemas
├── pkg/                    # reusable helpers: server/, queue/, cache/, metrics/, telegram/, util/, system/
├── migrations/             # SQL migrations: postgres/, sqlite/
├── tests/                  # bdd/, acceptance/, smoke/, torture/, performance/, warn_up/
├── docs/                   # human documentation; overview.md is the architecture reference
├── deployments/            # deployment and storage configuration
└── .ai/                    # this folder: context and rules for AI assistants
```

`contract/` is empty and `storage/` holds runtime artefacts (logs, profiles, the SQLite
file, generated data). Neither is a place for new code.

## Layering, and what it forbids

- `domain` — types only. No dependency on anything else in the project.
- `core` — interfaces and calculations. It must not import HTTP, SQL or exchange clients.
  A calculation that needs data takes it through an interface the caller implements.
- `internal/infrastructure` — the implementations of those interfaces: repositories,
  exchange adapters, caches, queues.
- `internal/ui` — entry points. Handlers translate between the transport layer and `core`;
  they hold no business logic worth testing on its own.
- `pkg` — code with no knowledge of this project's domain. If it mentions candlesticks, it
  belongs in `core` instead.

## Where new code goes

| Task | Place |
| --- | --- |
| New calculation or indicator | `core/<area>/`, with unit tests next to it and a BDD scenario in `tests/bdd/features/` |
| New REST endpoint | `api/rest/v1/openapi.yaml` first, regenerate, then a handler in `internal/ui/api/v1/impl/` |
| New exchange or data source | adapter in `internal/infrastructure/adapters/`, wired in `cmd/container/` |
| New scheduled job | `internal/ui/daemon/loader/` or a new daemon under `internal/ui/daemon/` |
| New page or asset | `internal/ui/web/static/`, route in `internal/ui/web/pages.go` |
| New domain type | `domain/`, and only there |

## Generated code

Two generators, both driven by files in `api/`:

- **REST** — `oapi-codegen` v2.5.0 produces `internal/ui/api/v1/spec/types.gen.go` and
  `server.gen.go` from `api/rest/v1/openapi.yaml`. Flags live in
  `internal/ui/api/v1/spec/generator.go`. Regenerate with the version that matches the
  header of the existing files; the `go install ...@v2.4.0` line in `generator.go` is stale.
- **gRPC** — `protoc` produces `internal/ui/api/grpc/action/` from `api/grpc/v1/*.proto`.

The contract is also embedded into the binary (`api/rest/v1/spec.go`) and served at
`/docs/api/openapi.yaml`, so the documentation on the site cannot drift from the file the
handlers were generated from.

## The web interface

`internal/ui/web/pages.go` embeds `static/` with `go:embed` and registers the routes:
`/` (landing), `/tools`, `/docs/api`, `/docs/api/openapi.yaml`, `/static/...`,
`/favicon.ico`, `/robots.txt`, `/sitemap.xml`, `/llms.txt`.

There is no build step and no bundler. The tools page loads Vue and PrimeVue as ES modules
from a CDN through an `importmap`; the landing page is static markup. Because the paths are
strings inside embedded files, a typo compiles fine and only breaks in the browser — that
is what `internal/ui/web/pages_test.go` exists for, and why it is worth extending rather
than working around.
