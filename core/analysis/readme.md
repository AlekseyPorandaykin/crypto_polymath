# Аналитика (`core/analysis`)

Рассчитываем аналитику на основе первичных индикаторов и другой аналитики (каскад MACD).

## Уровни расчёта

1. **CalculatorByIndicator** — по `domain.Indicator` (RSI, MACD Main Line, TrendByEMA, …)
2. **CalculatorByAnalytic** — по `Analytic` (MACD Signal Line, MACD Histogram)

## Публичные методы сервиса

| Метод | Вход | Пакетный `FindMany` |
|-------|------|---------------------|
| `CalculateByIndicator` | 1 индикатор | `FindManyByIndicator` |
| `AnalyticByIndicators` | []индикатор | да |
| `CalculateByAnalytic` | 1 analytic | делегирует в `CalculateByAnalytics` |
| `CalculateByAnalytics` | []analytic | через `OscillatorByAnalytics` |
| `OscillatorByAnalytics` | []analytic, имя осциллятора, depth | да |

## Использование в Calculator

`internal/ui/daemon/calculator/handleAnalytic` передаёт всю группу сообщений одной пары в `CalculateByAnalytics`.

## Тесты

```bash
go test ./core/analysis/... -v
go test ./core/analysis/... -bench=. -benchmem -run=^$
```

Unit-тесты: `service_test.go`, `contract_test.go`, `calculators/*_test.go`.
