# Core

Бизнес-логика приложения. Каждый подпакет отвечает за свою предметную область и не зависит от инфраструктуры (БД, HTTP, очереди).

## Пакеты

| Пакет | Описание |
|-------|----------|
| [`trading`](trading/readme.md) | Расчёты маржинальной торговли: усреднение, ликвидация, PnL, риски, динамический стоп-лосс |
| [`indicator`](indicator/readme.md) | Первичные технические индикаторы (MA, EMA, Trend, Stochastic и др.) по свечам |
| [`candle_indicator`](candle_indicator/readme.md) | Свечные индикаторы (Heiken Ashi) — производные OHLC-данные |
| [`analysis`](analysis/readme.md) | Аналитика второго уровня (MACD, RSI, TrendByMA/EMA) на основе индикаторов |
| [`exchange`](exchange/readme.md) | Загрузка и хранение информации о торговых парах бирж |
| [`candlestick`](candlestick/readme.md) | Загрузка, хранение и выдача свечей с бирж |
| [`price`](price/readme.md) | Загрузка и хранение текущих цен символов |

## Принципы

- **Чистая архитектура**: пакеты `core` зависят только от `domain` и стандартной библиотеки.
- **Инверсия зависимостей**: каждый пакет определяет интерфейсы (`Repository`, `ExchangeLoader`), реализация — в `internal/infrastructure`.
- **Тестируемость**: бизнес-логика покрыта unit, fuzz, golden, BDD (Cucumber/Godog) тестами.

## Тесты

```bash
# Все тесты core
make test-core

# С race detector
go test -race ./core/...

# Benchmarks
make bench-core

# BDD (Cucumber)
make test-bdd
```
