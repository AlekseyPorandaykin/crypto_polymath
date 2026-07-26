# Candlestick (`core/candlestick`)

Загрузка, хранение и выдача OHLCV-свечей с криптобирж.

## Назначение

Центральный сервис для работы со свечами: загружает исторические данные через адаптеры бирж, сохраняет в хранилище, предоставляет методы выборки с фильтрацией по exchange/symbol/unit/interval.

## Структура

```
candlestick/
├── contract.go      # Интерфейсы: Candlestick, ExchangeLoader, Saver; типы ExchangeDTO
├── service.go       # Реализация сервиса
└── repository.go    # Интерфейс Repository (CRUD для свечей)
```

## Интерфейс `Candlestick`

```go
type Candlestick interface {
    AddLoader(exchange string, loader ExchangeLoader)
    LoadCandlesticks(ctx, exchange, symbol, unit, interval) ([]domain.Candlestick, error)
    UpdateCandlesticks(ctx, exchange, symbol, unit, interval) ([]domain.Candlestick, error)
    Candlesticks(ctx, exchange, symbol, unit, interval, limit) ([]domain.Candlestick, error)
    SequenceCandlesticks(ctx, exchange, symbol, unit, interval, limit) ([]domain.Candlestick, error)
    CandlesticksToDate(ctx, exchange, symbol, unit, interval, limit, to) ([]domain.Candlestick, error)
    CandlesticksFromDate(ctx, exchange, symbol, unit, interval, limit, to) ([]domain.Candlestick, error)
    SequenceCandlesticksToDate(ctx, ...) ([]domain.Candlestick, error)
    DeleteOldRows(ctx, oldValueLimit) error
}
```

## Интерфейс `ExchangeLoader`

```go
type ExchangeLoader interface {
    LastMinuteCandlesticks(ctx, symbol, minutes) ([]ExchangeDTO, error)
    LastHourCandlesticks(ctx, symbol, hours) ([]ExchangeDTO, error)
    LastDayCandlesticks(ctx, symbol) ([]ExchangeDTO, error)
    LastWeekCandlesticks(ctx, symbol) ([]ExchangeDTO, error)
    LastMonthCandlesticks(ctx, symbol) ([]ExchangeDTO, error)
}
```

## Поддерживаемые таймфреймы

Конфигурация через `viper`: `candlestick.minutes` и `candlestick.hours`.

| Unit | Примеры интервалов |
|------|-------------------|
| Minute | 1, 3, 5, 15, 30 |
| Hour | 1, 2, 4 |
| Day | 1 |
| Week | 1 |
| Month | 1 |

## Ключевые особенности

- **Строгая последовательность** (`SequenceCandlesticks`): возвращает свечи только если нет пропусков во временном ряде, иначе — пустой результат.
- **Автоматическая очистка** (`DeleteOldRows`): удаляет устаревшие записи для экономии хранилища.
- **Конвертация**: `ExchangeDTO` (строковые цены) → `domain.Candlestick` (float64).
