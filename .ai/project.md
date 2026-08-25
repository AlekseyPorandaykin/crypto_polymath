# What Crypto Polymath is

A server platform for technical analysis of crypto markets. It pulls market data from
exchanges, computes indicators and derived analytics, stores the results in PostgreSQL and
serves them over a REST API, a gRPC service and a small web interface.

The service is built for automated collection across many symbols and timeframes at once —
not for hand-analysing a single chart. That bias explains most design decisions: data is
normalised into one shape for every exchange, calculations are driven by events rather than
by requests, and the public API answers the same way regardless of which venue the numbers
came from.

## Three processes, one binary

The binary is a cobra CLI (`main.go` → `cmd/`). Each process runs on its own and has its
own container in `docker-compose.yaml`:

| Process | Command | Responsibility |
| --- | --- | --- |
| Loader | `daemon loader` | Fetch prices, candlesticks and symbol metadata from exchanges; write them down; raise events |
| Calculator | `daemon calculator` | Consume events and compute indicators and analytics |
| API | `api external-v1` | Serve the REST API and the web pages |

All three build their dependencies through the shared DI container in `cmd/container`
(`go.uber.org/dig`): connections, repositories, services and exchange clients are
registered once at start-up.

## What it computes

- **Prices** — the latest price per exchange and symbol, all prices on one exchange, or one
  symbol across every exchange.
- **Candlesticks** — OHLCV on timeframes from minutes to months, plus Heiken Ashi
  smoothing (`core/candle_indicator`).
- **Indicators** — moving averages, RSI, MACD, stochastic and trend metrics
  (`core/indicator`), with a configurable lookback depth.
- **Analytics** — second-order metrics computed from indicators (`core/analysis`).
- **Trader calculators** — average entry price, position size from margin, liquidation
  price, distance to liquidation, unrealized and spot PnL, add-on simulation and a full
  risk snapshot (`core/trading`). These are pure functions: no exchange data involved.

## Exchanges

Binance, Bybit, OKX, Kraken, KuCoin, Gate.io, Bitget and MEXC deliver prices. **Bybit** is
the main source of candlesticks and futures symbol metadata; the other venues are used
mostly for price aggregation. The public site advertises six major exchanges — that number
is asserted by a test, so do not change the wording on the landing page casually.

Exchange clients come from the external `crypto_loader` and `crypto-exchanges` libraries;
adapters live in `internal/infrastructure/adapters`.

## Who talks to it

- **Visitors** of the landing page (`/`), the tools page (`/tools`) and the API
  documentation (`/docs/api`).
- **Integrations** calling `/api/v1/...`. Anonymous callers get 10 requests per minute per
  IP address; callers sending an `X-Token` header get the quota of their pricing plan.
- **The team**, through the contact form on the landing page, which posts to
  `POST /api/v1/contact`.

## What it is not

- Not a trading system: it never places orders and never asks for exchange API keys.
- Not investment advice — the site says so, and that wording should stay.
- Not a real-time feed: the API serves what the loader has already stored, so freshness is
  bounded by the loader's schedule.
