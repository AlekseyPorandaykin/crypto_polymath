# Промпт: стратегия Heiken Ashi + динамический Stop-Loss

Используй этот файл как контекст при вопросах о стратегии, бэктесте, SL и планах анализа.
Полные результаты: `docs/HeikenAshi.md`. Код: `core/trading/future_stop_loss.go`.

Связанные промпты: `prompts/trend-range-strategy.md` (тренд + границы диапазона, draft).

---

- **Crypto Polymath** — Go-сервис технического анализа крипторынка.
- Свечи: PostgreSQL / CSV в `storage/data/`.
- Heiken Ashi: `core/candle_indicator`, имя `HeikenAshi`.
- Фьючерсные расчёты (плечо, SL, PnL на марже): `core/trading.Future`.

---

## Стратегия (как должна работать)

### Таймфрейм

- Основной бэктест: **1 час**.
- Планируется анализ на **большом массиве** (месяцы данных, больше символов, возможно 1m для BTC/ETH).

### Сигнал входа

1. На **Heiken Ashi** появляется **доджи**: `body / range × 100 ≤ 10%` (или `open == close`).
2. В течение **3 свечей** после доджи формируются **2 HA-свечи подряд** в одном направлении:
   - 2× bullish (`close > open`) → **LONG**
   - 2× bearish (`close < open`) → **SHORT**
3. Вход на **open следующей реальной свечи** (не HA).
4. Одна позиция на символ; пока открыта — новые сигналы не обрабатываются.

### Stop-Loss (актуальная версия — динамический от волатильности)

**Не использовать** tight trailing SL с входа (HA low сразу) — доказано убыточно (WR ~24%).

**Рекомендуемая формула:**

```
volatility = (HA_high - HA_low) / close * 100   # range_pct — лучший предиктор

sl_price_%    = clamp(k_sl   × volatility, sl_floor=3%, sl_cap=15%)
trail_price_% = clamp(k_trail × volatility, trail_floor=2%, trail_cap=20%)

k_sl = 4.0, k_trail = 4.0   # DefaultStopLossCoefficients()
```

**Фазы SL:**

1. **До активации трейлинга:** фиксированный SL на `sl_price_%` от входа.
   - На марже: `initial_sl_margin_% = -sl_price_% × leverage`
2. **Активация трейлинга:** когда peak PnL на марже ≥ `trail_price_% × leverage`.
   - Сначала SL → breakeven (цена входа).
3. **После активации:** SL подтягивается за HA-свечой:
   - LONG: `stop = max(stop, HA_low_prev, entry)`
   - SHORT: `stop = min(stop, HA_high_prev, entry)`
4. **Выход:** касание SL внутри свечи (`low ≤ stop` / `high ≥ stop`), цена выхода с учётом гэпа.

### Плечо

- Бэктест: 3x–5x.
- **Лучший режим:** 5x + dynamic SL + **short only** + `sl_floor=5%`.
- Baseline 5x / fixed −10%/+10% маржи — **отклонён** (PF 0.63).

---

## Результаты бэктеста (кратко, storage/data)

| Режим | Total PnL* | PF | WR |
|-------|------------|-----|-----|
| Tight HA trail с входа | −235% | 0.84 | 24% |
| Fixed 5x −10%/+10% | −3 954% | 0.63 | 34% |
| Fixed 3x −25%/+20% | +243% | 1.04 | 50% |
| Dynamic 5x k=4 range | **+1 191%** | **1.10** | 51% |
| Dynamic 5x short floor 5% | **+4 198%** | **1.57** | 56% |

\*Сумма PnL% на марже по сделкам (каждая сделка = 1 unit margin). Не эквити-кривая.

**Период:** ~100 часов, 499 символов, медвежий фон (avg B&H −2.85%).

**Ключевые выводы:**

- SL должен **масштабироваться с волатильностью** (теория подтверждена).
- **range_pct** HA-свечи лучше ATR и комбинированной mkt_vol.
- **Short** >> **Long** на этом периоде.
- После активации трейлинга WR ~**64–92%**; основной убыток — сделки до активации.
- Низкая волатильность (Q1): нужен **sl_floor**, иначе SL слишком узкий.

---

## API в коде (`core/trading.Future`)

```go
coef := trading.DefaultStopLossCoefficients()
f := trading.Future{Leverage: 5}

// Волатильность
rangePct := trading.VolatilityRangePercent(haHigh, haLow, close)
vol := trading.VolatilitySnapshot{RangePct: rangePct, ATRPct: atr, MktVol: mkt}

// SL
sl := f.DynamicStopLoss(side, entry, vol, coef, trading.VolatilityRange)

// Трейлинг
f.ShouldActivateTrailing(peakMarginPct, sl.TrailActivateMarginPct)
upd := f.UpdateTrailingStop(side, entry, stop, haLow, haHigh, activated)

// Выход
trading.IsStopHit(side, stop, low, high)
exit := trading.StopExitPrice(side, stop, open)
pnl := f.MarginPnLPercent(side, entry, exit)

// Доджи
trading.IsHeikenAshiDoji(haOpen, haClose, haHigh, haLow, coef.DojiBodyPct)
```

---

## Что отвечать на типичные вопросы

### «Как работает стратегия?»
→ Доджи HA → 2 свечи в новом направлении → вход → dynamic SL → трейлинг после порога профита.

### «Какие параметры SL?»
→ `DefaultStopLossCoefficients()`: k_sl=4, k_trail=4, range_pct, floor 3–5%. См. `docs/HeikenAshi.md`.

### «Почему long убыточен?»
→ Медвежий период данных; 44% ложных разворотов; BTC фактор. Short-only или фильтр BTC>MA24.

### «Что дальше?»
→ Большой массив данных, walk-forward, комиссии, бэктест-движок, top-N ликвидность.

### «Где данные?»
→ `storage/data/crypto_app_crypto_polymath_candlestick.csv`, индикаторы HA в `storage/data/crypto_app_crypto_poandlestick_indicators.csv`.

---

## Ограничения модели

- Нет комиссий, funding, slippage.
- Isolated margin, одна позиция на символ.
- HA warmup: нужна полная история до первой свечи окна.
- Предварительная статистика на ~4 днях — не финальная для продакшена.

---

## Файлы для чтения

1. `docs/HeikenAshi.md` — полный отчёт
2. `core/trading/future_stop_loss.go` — расчёты SL
3. `core/trading/contract.go` — типы `StopLossCoefficients`, `VolatilitySnapshot`, `DynamicStopLoss`
4. `core/candle_indicator/calculators/heiken_ashi.go` — формула HA
5. `prompts/heiken-ashi-strategy.md` — этот файл
