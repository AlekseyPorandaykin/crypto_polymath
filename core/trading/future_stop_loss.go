package trading

import "math"

// DefaultStopLossCoefficients — рекомендуемые коэффициенты по результатам бэктеста (см. docs/HeikenAshi.md).
func DefaultStopLossCoefficients() StopLossCoefficients {
	return StopLossCoefficients{
		KSL:           4.0,
		KTrail:        4.0,
		SLFloorPct:    3.0,
		TrailFloorPct: 2.0,
		DojiBodyPct:   10.0,
		SLCapPct:      15.0,
		TrailCapPct:   20.0,
	}
}

// Measure возвращает выбранную метрику волатильности.
func (v VolatilitySnapshot) Measure(kind VolatilityMeasure) float64 {
	switch kind {
	case VolatilityATR:
		return v.ATRPct
	case VolatilityCombined:
		return v.MktVol
	default:
		return v.RangePct
	}
}

// VolatilityRangePercent — колебание свечи в % от цены: (high−low)/close × 100.
func VolatilityRangePercent(high, low, close float64) float64 {
	if close <= 0 {
		return 0
	}
	return (high - low) / close * 100
}

// IsHeikenAshiDoji — доджи HA: тело ≤ threshold% от диапазона свечи.
func IsHeikenAshiDoji(open, close, high, low, dojiThresholdPct float64) bool {
	size := high - low
	if size <= 0 {
		return true
	}
	body := math.Abs(close - open)
	if body == 0 {
		return true
	}
	return body/size*100 <= dojiThresholdPct
}

// PriceChangePercent — изменение цены от входа в % (положительное = в сторону профита).
func (f Future) PriceChangePercent(side Side, entry, price float64) float64 {
	if entry <= 0 || price <= 0 {
		return 0
	}
	switch side {
	case Long:
		return (price/entry - 1) * 100
	case Short:
		return (entry/price - 1) * 100
	default:
		return 0
	}
}

// MarginPnLPercent — PnL в % от залога с учётом плеча f.Leverage.
func (f Future) MarginPnLPercent(side Side, entry, price float64) float64 {
	return f.PriceChangePercent(side, entry, price) * f.Leverage
}

// PriceForMarginPnLPercent — цена при заданном PnL на марже (marginPnL% = priceChange% × leverage).
func (f Future) PriceForMarginPnLPercent(side Side, entry, marginPnLPct float64) float64 {
	if entry <= 0 || f.Leverage <= 0 {
		return 0
	}
	priceChangePct := marginPnLPct / f.Leverage
	switch side {
	case Long:
		return entry * (1 + priceChangePct/100)
	case Short:
		return entry * (1 - priceChangePct/100)
	default:
		return 0
	}
}

// ClampVolatilityPct ограничивает и нормализует значение волатильности.
func ClampVolatilityPct(v float64) float64 {
	if v < 0.1 {
		return 0.1
	}
	return v
}

func clampPct(v, floor, cap float64) float64 {
	if floor > 0 && v < floor {
		v = floor
	}
	if cap > 0 && v > cap {
		v = cap
	}
	return v
}

// DynamicStopLoss — SL и порог трейлинга от волатильности свечи.
//
//	sl_price_%    = clamp(KSL × vol, SLFloorPct, SLCapPct)
//	trail_price_% = clamp(KTrail × vol, TrailFloorPct, TrailCapPct)
//
// vol берётся из VolatilitySnapshot.Measure(measure); по умолчанию VolatilityRange.
func (f Future) DynamicStopLoss(side Side, entry float64, vol VolatilitySnapshot, coef StopLossCoefficients, measure VolatilityMeasure) DynamicStopLoss {
	v := ClampVolatilityPct(vol.Measure(measure))

	slPct := coef.KSL * v
	trailPct := coef.KTrail * v
	slPct = clampPct(slPct, coef.SLFloorPct, coef.SLCapPct)
	trailPct = clampPct(trailPct, coef.TrailFloorPct, coef.TrailCapPct)

	initialSLMargin := -slPct * f.Leverage
	trailMargin := trailPct * f.Leverage

	result := DynamicStopLoss{
		VolatilityUsed:         v,
		SLPricePct:             slPct,
		TrailActivatePricePct:  trailPct,
		InitialSLMarginPct:     initialSLMargin,
		TrailActivateMarginPct: trailMargin,
	}

	result.InitialStopPrice = f.PriceForMarginPnLPercent(side, entry, initialSLMargin)

	switch side {
	case Long:
		result.TrailActivatePrice = entry * (1 + trailPct/100)
	case Short:
		result.TrailActivatePrice = entry * (1 - trailPct/100)
	}

	return result
}

// UpdateTrailingStop — обновляет трейлинг-SL синхронно с HA-свечой.
//
// До активации (peakMargin < trailActivateMargin) возвращает currentStop без изменений.
// После активации: long — max(stop, haLow, entry); short — min(stop, haHigh, entry).
func (f Future) UpdateTrailingStop(side Side, entry, currentStop, haLow, haHigh float64, trailActivated bool) TrailingStopUpdate {
	if !trailActivated {
		return TrailingStopUpdate{StopPrice: currentStop, TrailActivated: false}
	}

	switch side {
	case Long:
		stop := currentStop
		if entry > stop {
			stop = entry
		}
		if haLow > stop {
			stop = haLow
		}
		return TrailingStopUpdate{StopPrice: stop, TrailActivated: true}
	case Short:
		stop := currentStop
		if entry < stop {
			stop = entry
		}
		if haHigh < stop {
			stop = haHigh
		}
		return TrailingStopUpdate{StopPrice: stop, TrailActivated: true}
	default:
		return TrailingStopUpdate{StopPrice: currentStop, TrailActivated: trailActivated}
	}
}

// ShouldActivateTrailing — достигнут ли порог активации трейлинга по пику PnL на марже.
func (f Future) ShouldActivateTrailing(peakMarginPnLPct, trailActivateMarginPct float64) bool {
	return peakMarginPnLPct >= trailActivateMarginPct
}

// IsStopHit — сработал ли SL внутри свечи (по low/high).
func IsStopHit(side Side, stop, candleLow, candleHigh float64) bool {
	switch side {
	case Long:
		return candleLow <= stop
	case Short:
		return candleHigh >= stop
	default:
		return false
	}
}

// StopExitPrice — цена выхода по стопу с учётом гэпа через уровень.
func StopExitPrice(side Side, stop, candleOpen float64) float64 {
	switch side {
	case Long:
		if candleOpen < stop {
			return candleOpen
		}
		return stop
	case Short:
		if candleOpen > stop {
			return candleOpen
		}
		return stop
	default:
		return stop
	}
}
