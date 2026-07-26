# Price (`core/price`)

Загрузка и хранение текущих цен символов с криптобирж.

## Назначение

Сервис отвечает за получение актуальных цен (last price) по торговым парам, их хранение в БД и предоставление через запросы по exchange/symbol.

## Структура

```
price/
├── contract.go      # Интерфейсы: Price, ExchangeLoader; типы ExchangeDTO
├── service.go       # Реализация сервиса
└── repository.go    # Интерфейс Repository
```

## Интерфейс `Price`

```go
type Price interface {
    AddLoader(exchange string, loader ExchangeLoader)
    LoadPrices(ctx, exchange) ([]domain.Price, error)
    Save(ctx, ...domain.Price) error
    LastPrice(ctx, exchange, symbol) (*domain.Price, error)
    LastPricesByExchange(ctx, exchange) ([]domain.Price, error)
    LastPricesBySymbol(ctx, symbol) ([]domain.Price, error)
    DeleteOldRaws(ctx, exchange, to) error
}
```

## Интерфейс `ExchangeLoader`

```go
type ExchangeLoader interface {
    Prices(ctx) ([]ExchangeDTO, error)
    Price(ctx, symbol) (ExchangeDTO, error)
}
```

## Логика работы

1. **LoadPrices** — загружает все цены с биржи → сохраняет в БД → удаляет предыдущие записи.
2. **LastPrice** — получение последней сохранённой цены.
3. **LastPricesByExchange** / **LastPricesBySymbol** — групповые выборки.
4. **Retry с backoff** — сохранение использует exponential backoff для устойчивости к временным ошибкам БД.

## Ключевые особенности

- **Atomic update**: `LoadPrices` атомарно заменяет данные (save → delete old).
- **Multi-exchange**: поддержка нескольких бирж через `AddLoader`.
- **Конвертация**: строковая цена из биржевого API → `float64` в domain.
