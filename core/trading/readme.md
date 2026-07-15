# Trading (`core/trading`)

Расчёты для маржинальной торговли: усреднение позиции, ликвидация, PnL и риски.
Модель — **изолированная маржа USDT-M** (одна позиция, залог в USDT).

## Сценарий: лонг → падение → докупка

1. Открыт лонг: объём `V₁`, вход `P₁`, залог `M₁`, плечо `L`.
2. Цена падает до `P₂`, докупаем на залог `M₂` (или объём `V₂`).
3. Пересчитываются:
   - средняя цена входа;
   - общий объём и залог;
   - цена ликвидации;
   - PnL и запас до ликвидации.

Для **шорта** логика зеркальная: цена растёт → добавляем к шорту → средняя цена входа **растёт**.

## Формулы

### Средняя цена входа

```
P_avg = (V₁×P₁ + V₂×P₂) / (V₁ + V₂)
```

```go
NewAvgEntryPrice(entryVolume, entryPrice, newVolume, newPrice)
```

### Докупка на сумму залога

```
V₂ = M₂ × L / P₂
P_avg = NewAvgEntryPrice(V₁, P₁, V₂, P₂)
```

```go
Future{Leverage: leverage}.NewAvgEntryPriceBySum(entryVolume, entryPrice, sum, newPrice)
```

### Нереализованный PnL

| Сторона | Формула |
|---------|---------|
| Long | `V × (P_mark − P_entry)` |
| Short | `V × (P_entry − P_mark)` |

### Цена ликвидации (изолированная позиция)

Условие: `Margin + PnL = Notional × MMR`.

| Сторона | Формула |
|---------|---------|
| Long | `P_liq = (M − V×E) / (V×(MMR − 1))` |
| Short | `P_liq = (M + V×E) / (V×(1 + MMR))` |

- `M` — залог (USDT)
- `V` — объём позиции
- `E` — средняя цена входа
- `MMR` — maintenance margin rate (например `0.005` = 0.5%)

```go
LiquidationPrice(side, volume, entry, margin, maintenanceMarginRate)
```

### Запас до ликвидации (%)

| Сторона | Формула |
|---------|---------|
| Long | `(P_mark − P_liq) / P_mark × 100` |
| Short | `(P_liq − P_mark) / P_mark × 100` |

### Эффективное плечо

```
L_eff = (V × P_mark) / M
```

## API

| Функция | Назначение |
|---------|------------|
| `SimulateAddOn` | Полный снимок до/после докупки |
| `RiskAtPrice` | Риски позиции при текущей цене |
| `AddOnResult` | Entry delta, ликвидация, PnL, distance to liq |
| `RiskSnapshot` | PnL, margin usage, effective leverage |
| `Future` | Расчёты, которым нужно плечо (`Leverage`): методы `NewAvgEntryPriceBySum`, `VolumeFromMargin`, `ResolveAddOnVolume`, **динамический SL** (`DynamicStopLoss`, `UpdateTrailingStop`) |

Подробнее о стратегии Heiken Ashi и бэктесте: [`docs/HeikenAshi.md`](../../docs/HeikenAshi.md).

## Аргументы

### `Side` — направление позиции

| Значение | Смысл |
|----------|--------|
| `Long` (= 1) | Лонг — прибыль при росте цены |
| `Short` (= -1) | Шорт — прибыль при падении цены |

### `Position` — текущая позиция

| Поле | Тип | Единица | Описание |
|------|-----|---------|----------|
| `Side` | `Side` | — | Long или Short |
| `Volume` | `float64` | BTC, ETH, … | Размер позиции в базовом активе |
| `EntryPrice` | `float64` | USDT | Средняя цена входа |
| `Margin` | `float64` | USDT | Залог (isolated wallet) |
| `Leverage` | `float64` | × | Плечо для расчёта объёма из суммы докупки |

### `AddOn` — докупка

| Поле | Тип | Единица | Описание |
|------|-----|---------|----------|
| `Price` | `float64` | USDT | Цена, по которой докупаем |
| `Volume` | `float64` | базовый актив | Доп. объём; если `> 0` — используется напрямую |
| `Margin` | `float64` | USDT | Доп. залог; при `Volume = 0` объём = `Margin × Leverage / Price` |

Если задан `Volume > 0`, поле `Margin` всё равно **прибавляется** к залогу позиции.

### `maintenanceMarginRate` (MMR)

Общий параметр в `LiquidationPrice`, `SimulateAddOn`, `RiskAtPrice`.

| | |
|---|---|
| Тип | `float64` |
| Диапазон | `0 < MMR < 1` |
| Пример | `0.005` = 0.5% |
| Смысл | Доля номинала как maintenance margin на бирже |

### Функции — входные аргументы

#### `NewAvgEntryPrice(entryVolume, entryPrice, newVolume, newPrice)`

| Аргумент | Смысл |
|----------|--------|
| `entryVolume` | Текущий объём V₁ |
| `entryPrice` | Текущий вход P₁ |
| `newVolume` | Объём докупки V₂ |
| `newPrice` | Цена докупки P₂ |

Возврат: новая средняя цена входа (USDT).

### `Future` — расчёты с плечом

Методы, которым для вычисления нужно плечо, вынесены на `Future{Leverage: L}`, а не принимают
`leverage` отдельным аргументом в каждой функции.

#### `Future{Leverage}.NewAvgEntryPriceBySum(entryVolume, entryPrice, sum, newPrice)`

| Аргумент | Смысл |
|----------|--------|
| `Leverage` (поле `Future`) | Плечо L |
| `entryVolume` | V₁ |
| `entryPrice` | P₁ |
| `sum` | Сумма залога докупки (USDT) |
| `newPrice` | Цена докупки P₂ |

Внутри: `newVolume = sum × leverage / newPrice`.

#### `Future{Leverage}.VolumeFromMargin(margin, price)`

`volume = margin × leverage / price`

#### `Future{Leverage}.ResolveAddOnVolume(add AddOn)`

Объём докупки: явный `add.Volume`, либо `VolumeFromMargin(add.Margin, add.Price)` с плечом `Future.Leverage`.
Используется внутри `SimulateAddOn` с плечом текущей позиции (`Future{Leverage: pos.Leverage}`).

#### `Notional(volume, price)`

`volume × price` — номинал в USDT.

#### `UnrealizedPnL(side, volume, entry, mark)`

| Аргумент | Смысл |
|----------|--------|
| `side` | `Long` / `Short` |
| `volume` | Объём позиции |
| `entry` | Средняя цена входа |
| `mark` | Рыночная цена |

Возврат: PnL в USDT (отрицательный = убыток).

#### `LiquidationPrice(side, volume, entry, margin, maintenanceMarginRate)`

| Аргумент | Смысл |
|----------|--------|
| `side` | `Long` / `Short` |
| `volume` | V |
| `entry` | E — средний вход |
| `margin` | M — залог |
| `maintenanceMarginRate` | MMR |

Возврат: цена ликвидации (USDT).

#### `EffectiveLeverage(volume, price, margin)`

`(volume × price) / margin`

#### `DistanceToLiquidationPercent(side, mark, liquidation)`

Запас до ликвидации в % от `mark`. Отрицательное значение — цена уже за уровнем ликвидации.

#### `SimulateAddOn(pos, add, maintenanceMarginRate)`

| Аргумент | Тип | Описание |
|----------|-----|----------|
| `pos` | `Position` | Позиция до докупки |
| `add` | `AddOn` | Параметры докупки |
| `maintenanceMarginRate` | `float64` | MMR |

Возврат: `AddOnResult`.

#### `RiskAtPrice(pos, mark, maintenanceMarginRate)`

| Аргумент | Тип | Описание |
|----------|-----|----------|
| `pos` | `Position` | Текущая позиция |
| `mark` | `float64` | Рыночная цена для оценки |
| `maintenanceMarginRate` | `float64` | MMR |

Возврат: `RiskSnapshot`.

### `AddOnResult` — поля результата докупки

| Поле | Единица | Смысл |
|------|---------|--------|
| `Before` | `Position` | Позиция до докупки |
| `After` | `Position` | Позиция после |
| `VolumeAdded` | базовый актив | Добавленный объём |
| `EntryDelta` | USDT | `After.Entry − Before.Entry` |
| `EntryDeltaPercent` | % | Изменение входа относительно старого |
| `LiquidationBefore` | USDT | Ликвидация до |
| `LiquidationAfter` | USDT | Ликвидация после |
| `LiquidationDelta` | USDT | `After − Before` |
| `MarginAdded` | USDT | = `add.Margin` |
| `NotionalAfter` | USDT | `After.Volume × add.Price` |
| `EffectiveLeverage` | × | Плечо после докупки |
| `MaintenanceMargin` | USDT | `Notional × MMR` |
| `UnrealizedPnLAtPrice` | USDT | PnL на цене докупки |
| `PnLPercentOnMargin` | % | PnL / залог × 100 |
| `DistanceToLiquidationPct` | % | Запас до ликвидации на цене докупки |
| `BreakEvenPrice` | USDT | Безубыток (= `After.EntryPrice`) |

**Лонг при падении:** `EntryDelta < 0`, `LiquidationDelta < 0`.

**Шорт при росте:** `EntryDelta > 0`, `LiquidationDelta > 0`.

### `RiskSnapshot` — поля риска

| Поле | Единица | Смысл |
|------|---------|--------|
| `MarkPrice` | USDT | Цена, на которой считали |
| `EntryPrice` | USDT | Средний вход |
| `LiquidationPrice` | USDT | Ликвидация |
| `UnrealizedPnL` | USDT | Текущий PnL |
| `PnLPercentOnMargin` | % | PnL от залога |
| `Notional` | USDT | `Volume × MarkPrice` |
| `EffectiveLeverage` | × | Фактическое плечо |
| `MaintenanceMargin` | USDT | Требуемая maintenance margin |
| `MarginUsagePercent` | % | `MaintenanceMargin / (Margin + PnL) × 100` |
| `DistanceToLiquidationPct` | % | Запас до ликвидации |

### Пример: лонг, усреднение вниз

```go
pos := trading.Position{
    Side:       trading.Long,
    Volume:     1,
    EntryPrice: 100_000,
    Margin:     10_000,
    Leverage:   10,
}

result := trading.SimulateAddOn(pos, trading.AddOn{
    Price:  90_000,
    Margin: 10_000,
}, 0.005)

// result.After.EntryPrice          — новая средняя (ниже 100k)
// result.EntryDeltaPercent         — на сколько % сместился вход
// result.LiquidationAfter          — новая ликвидация (ниже, чем до докупки)
// result.DistanceToLiquidationPct — запас до ликвидации на цене докупки
// result.UnrealizedPnLAtPrice      — PnL сразу после усреднения
```

### Пример: шорт, добавление при росте

```go
pos := trading.Position{
    Side:       trading.Short,
    Volume:     1,
    EntryPrice: 100_000,
    Margin:     10_000,
    Leverage:   10,
}

result := trading.SimulateAddOn(pos, trading.AddOn{
    Price:  110_000,
    Margin: 10_000,
}, 0.005)

// result.After.EntryPrice   — выше 100k
// result.LiquidationAfter   — выше (шорт ликвидируется при росте)
```

## Что меняется при докупке

| Параметр | Long (цена ↓) | Short (цена ↑) |
|----------|---------------|----------------|
| Средний вход | **Снижается** | **Растёт** |
| Объём | Растёт | Растёт |
| Залог | Растёт | Растёт |
| Ликвидация | Смещается **вниз** | Смещается **вверх** |
| Безубыток | = новый `EntryPrice` | = новый `EntryPrice` |

## Риски

- **Больший объём** — усреднение увеличивает экспозицию; откат к безубытку легче, но убыток при дальнейшем движении против позиции больше.
- **Margin usage** — доля maintenance margin от equity (`RiskSnapshot.MarginUsagePercent`); чем выше, тем ближе к ликвидации.
- **MMR** — зависит от биржи и размера позиции; передаётся параметром (в тестах `0.005`).
- **Комиссии, funding, cross-margin** — в текущей модели не учитываются.

## Тесты и benchmarks

```bash
go test ./core/trading/... -v
go test ./core/trading/... -bench=. -benchmem -run=^$
```

## Ограничения модели

- Только **isolated margin**, одна позиция.
- `MMR` задаётся вручную (не таблица tier биржи).
- Нет учёта fees, slippage, funding rate.
- Для cross-margin и hedge mode нужна отдельная модель.

## Динамический Stop-Loss (Heiken Ashi стратегия)

Расчёты в `future_stop_loss.go` на `Future{Leverage}`:

| Метод / тип | Назначение |
|-------------|------------|
| `DefaultStopLossCoefficients()` | Рекомендуемые k_sl, k_trail, floor/cap |
| `VolatilitySnapshot` | range_pct, atr_pct, mkt_vol на свече |
| `DynamicStopLoss` | Начальный SL и порог трейлинга от vol |
| `UpdateTrailingStop` | Подтягивание SL за HA-свечой |
| `MarginPnLPercent` / `PriceForMarginPnLPercent` | PnL и цена по % маржи |
| `IsHeikenAshiDoji` | Детекция доджи |

```go
f := trading.Future{Leverage: 5}
coef := trading.DefaultStopLossCoefficients()
vol := trading.VolatilitySnapshot{RangePct: trading.VolatilityRangePercent(haH, haL, close)}
sl := f.DynamicStopLoss(trading.Long, entry, vol, coef, trading.VolatilityRange)
```

Документация и результаты бэктеста: `docs/HeikenAshi.md`, промпт для AI: `prompts/heiken-ashi-strategy.md`.
