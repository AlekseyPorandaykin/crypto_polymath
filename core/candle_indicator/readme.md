# Candle Indicator (`core/candle_indicator`)

Рассчитывает свечные индикаторы — производные свечи (Heiken Ashi и др.), сохраняя результат как OHLC-запись с привязкой к исходной свече.

## Архитектура

```
candle_indicator/
├── contract.go          # Типы: Indicator (OHLC + мета), интерфейс Calculator
├── service.go           # Сервис: оркестрация расчёта
├── repository.go        # Repository: хранение свечных индикаторов
└── calculators/
    └── heiken_ashi.go   # Heiken Ashi (сглаженные свечи)
```

## Тип `Indicator`

```go
type Indicator struct {
    Name       string
    Exchange   string
    Symbol     string
    Unit       domain.Unit
    Interval   int
    StartTime  time.Time
    OpenPrice  float64
    HighPrice  float64
    LowPrice   float64
    ClosePrice float64
}
```

Вспомогательные методы: `SizeBody()`, `Size()`, `SizeBodyInPercent()`, `CloseLocation()`, `OpenLocation()`, `IsUp()`, `IsDown()`, `Direction()`, `PrevStartTime()`.

## Интерфейс `Calculator`

```go
type Calculator interface {
    Name() string
    Calculate(ctx context.Context, candle domain.Candlestick) (*Indicator, error)
}
```

Каждый калькулятор принимает одну свечу и может обращаться к хранилищу предыдущих данных (для Heiken Ashi — предыдущий HA).

## Калькуляторы

| Калькулятор | Описание |
|-------------|----------|
| `HeikenAshi` | Сглаженные свечи: `Close = (O+H+L+C)/4`, `Open = (prevHA.Close + prevHA.Open)/2` |

## Использование

Сервис подключает калькуляторы через `AddCalculator()`, затем при получении свечей автоматически рассчитывает и сохраняет свечные индикаторы.

## Тесты

```bash
go test ./core/candle_indicator/... -v
go test ./core/candle_indicator/... -bench=. -benchmem -run=^$
```

Unit-тесты: `contract_test.go`, `service_test.go`, benchmarks: `*_bench_test.go`.
