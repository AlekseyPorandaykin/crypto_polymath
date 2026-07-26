# Exchange (`core/exchange`)

Загрузка и хранение информации о торговых парах (символах) криптобирж.

## Назначение

Пакет управляет справочником символов: загружает актуальные данные с бирж через адаптеры, сохраняет в хранилище и предоставляет методы поиска.

## Структура

```
exchange/
├── contract.go      # Интерфейсы Exchange, ExternalLoader, типы SymbolInfoDTO, SymbolCategory
├── service.go       # Реализация сервиса
├── repository.go    # Интерфейс Repository
└── service_test.go  # Unit-тесты
```

## Интерфейс `Exchange`

```go
type Exchange interface {
    AddLoader(exchangeName string, loader ExternalLoader)
    LoadSymbolInfo(ctx context.Context, exchangeName string) ([]domain.SymbolInfo, error)
    SymbolInfo(ctx context.Context, exchangeName, symbol string) (*domain.SymbolInfo, error)
    SymbolInfoByCategory(ctx context.Context, exchange, category string) ([]domain.SymbolInfo, error)
}
```

## Категории символов

| Константа | Значение |
|-----------|----------|
| `SymbolCategorySpot` | `"spot"` |
| `SymbolCategoryFuture` | `"future"` |
| `SymbolCategoryOther` | `"other"` |

## Данные `SymbolInfoDTO`

| Поле | Описание |
|------|----------|
| `Symbol` | Тикер (BTCUSDT, ETHUSDT) |
| `Exchange` | Название биржи |
| `BaseAsset` | Базовый актив (BTC, ETH) |
| `QuoteAsset` | Котируемый актив (USDT) |
| `Category` | Spot / Future / Other |
| `FundingRate` | Текущий funding rate |
| `NextFundingTime` | Время следующего funding |

## Логика

1. `LoadSymbolInfo` — загружает с биржи, сохраняет в БД, удаляет устаревшие записи.
2. `SymbolInfo` — поиск по exchange + symbol; если не найден — пытается разобрать symbol по известным quote-активам.
3. `SymbolInfoByCategory` — фильтрация по категории (spot/future).

## Тесты

```bash
go test ./core/exchange/... -v
```
