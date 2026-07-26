# Аналитика (`core/analysis`)

Расчёт аналитики второго уровня на основе первичных индикаторов и каскадных зависимостей (MACD Signal → MACD Histogram).

## Архитектура

```
analysis/
├── contract.go              # Интерфейсы, типы
├── service.go               # Сервис: CalculateByIndicator, CalculateByAnalytics
├── repository.go            # Repository
├── readme.md
└── calculators/             # Конкретные аналитические калькуляторы
    ├── macd_main_line.go    # MACD Main Line = EMA(12) − EMA(26) (через индикаторы)
    ├── macd_signal_line.go  # MACD Signal = EMA(9) от MACD Main Line
    ├── macd_histogram.go    # MACD Histogram = Main − Signal
    ├── rsi.go               # Relative Strength Index
    ├── trend_by_ma.go       # Тренд по MA (TrendByMA)
    ├── trend_by_ema.go      # Тренд по EMA (TrendByEMA)
    ├── stochastic_signal_line.go  # %D (сигнальная линия стохастика)
    ├── candle_by_ma.go      # Отношение свечи к MA (RatioCandleToMA)
    └── candle_by_ema.go     # Отношение свечи к EMA (RatioCandleToEMA)
```

## Уровни расчёта

1. **CalculatorByIndicator** — принимает `[]domain.Indicator`, возвращает аналитику  
   Примеры: MACD Main Line, RSI, TrendByMA, TrendByEMA

2. **CalculatorByAnalytic** — принимает `[]Analytic`, возвращает аналитику  
   Примеры: MACD Signal Line, MACD Histogram, Stochastic Signal Line

Это позволяет строить каскады: индикаторы → аналитика 1-го уровня → аналитика 2-го уровня.

## Публичные методы сервиса

| Метод | Вход | Описание |
|-------|------|----------|
| `CalculateByIndicator` | 1 индикатор | Рассчитать аналитику по одному индикатору |
| `AnalyticByIndicators` | `[]Indicator` | Пакетный расчёт по группе индикаторов |
| `CalculateByAnalytic` | 1 Analytic | Рассчитать зависимую аналитику |
| `CalculateByAnalytics` | `[]Analytic` | Пакетный расчёт каскада |
| `OscillatorByAnalytics` | `[]Analytic`, name, depth | Расчёт осциллятора |

## Тесты

```bash
go test ./core/analysis/... -v
go test ./core/analysis/... -bench=. -benchmem -run=^$
```

Unit-тесты: `service_test.go`, `contract_test.go`, `calculators/*_test.go`, `calculators/helpers_test.go`.
