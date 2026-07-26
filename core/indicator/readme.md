# Индикаторы (`core/indicator`)

Сервис расчёта первичных (простых) технических индикаторов на основе свечей.

## Архитектура

```
indicator/
├── contract.go          # Интерфейсы Indicator, Candlestick
├── service.go           # Реализация: расчёт, кэш, персистентность
├── repository.go        # Интерфейс Repository (хранение индикаторов)
└── calculator/          # Подпакет с конкретными калькуляторами
    ├── contract.go      # Интерфейс PrimaryIndicatorCalculator
    ├── ma.go            # Simple Moving Average (MA)
    ├── ema.go           # Exponential Moving Average (EMA)
    ├── trend.go         # Определение тренда (Upward/Downward/Flat)
    ├── stochastic_main_line.go  # Стохастик (основная линия)
    ├── price_changes.go         # Изменение цены (%)
    ├── volatility_candle_percent.go  # Волатильность свечи (%)
    └── type_candle.go           # Тип свечи (форма, тело, тени)
```

## Интерфейс `PrimaryIndicatorCalculator`

```go
type PrimaryIndicatorCalculator interface {
    Name() string
    SupportDepth(depth int) bool
    SupportInterval(interval int) bool
    Calculate(candlesticks []domain.Candlestick) *domain.Indicator
}
```

Каждый калькулятор принимает слайс свечей и возвращает один `*domain.Indicator` (или `nil`, если данных недостаточно).

## Доступные индикаторы

| Калькулятор | Имя | Описание |
|-------------|-----|----------|
| `MA` | `"MA"` | Простая скользящая средняя по close |
| `EMA` | `"EMA"` | Экспоненциальная скользящая средняя |
| `Trend` | `"Trend"` | Тренд: +1 (up), -1 (down), 0 (flat) |
| `StochasticMainLine` | `"Stochastic"` | %K Стохастика (0–100) |
| `PriceChanges` | `"PriceChanges"` | Изменение цены в % за период |
| `VolatilityCandlePercent` | `"VolatilityCandlePercent"` | Диапазон свечи / close × 100 |
| `TypeCandle` | `"TypeCandle"` | Классификация свечи (форма/паттерн) |

## Сервис `Indicator`

Оркестрирует расчёт: подгружает свечи из `Candlestick`-репозитория, вызывает калькуляторы, сохраняет результаты через `Repository`.

Основные методы:

| Метод | Описание |
|-------|----------|
| `CalcIndicators` | Расчёт всех индикаторов для пары за последнюю свечу |
| `CalcIndicatorsByCandlestick` | То же по конкретной свече |
| `CalculateLastSequence` | Расчёт последовательности индикаторов за N свечей |
| `Indicators` | Получение из БД ранее сохранённых индикаторов |
| `LastSequenceToDate` | Последовательность индикаторов до указанной даты |

## Тесты

```bash
go test ./core/indicator/... -v
go test ./core/indicator/... -bench=. -benchmem -run=^$
go test ./core/indicator/calculator/... -fuzz=Fuzz -fuzztime=10s
```

Unit-тесты: `service_test.go`, `calculator/*_test.go`, `calculator/*_fuzz_test.go`.
