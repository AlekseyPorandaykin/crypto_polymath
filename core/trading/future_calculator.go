package trading

import "math"

// NewAvgEntryPrice — средневзвешенная цена входа после докупки.
// V1, P1 — текущая позиция; V2, P2 — докупка.
// P_avg = (V1×P1 + V2×P2) / (V1 + V2)
func (f Future) NewAvgEntryPrice(entryVolume, entryPrice, newVolume, newPrice float64) float64 {
	total := entryVolume + newVolume
	if total <= 0 {
		return 0
	}
	return (entryVolume*entryPrice + newVolume*newPrice) / total
}

// NewAvgEntryPriceBySum — докупка на фиксированную сумму залога с плечом f.Leverage.
// newVolume = sum × leverage / newPrice
func (f Future) NewAvgEntryPriceBySum(entryVolume, entryPrice, sum, newPrice float64) float64 {
	if newPrice <= 0 || f.Leverage <= 0 {
		return entryPrice
	}
	return f.NewAvgEntryPrice(entryVolume, entryPrice, sum*f.Leverage/newPrice, newPrice)
}

// VolumeFromMargin — объём позиции из залога и плеча f.Leverage.
// volume = margin × leverage / price
func (f Future) VolumeFromMargin(margin, price float64) float64 {
	if price <= 0 || f.Leverage <= 0 {
		return 0
	}
	return margin * f.Leverage / price
}

// Notional — номинал позиции: volume × price.
func Notional(volume, price float64) float64 {
	return volume * price
}

// UnrealizedPnL — нереализованный PnL при цене mark (USDT).
// По объёму: Long volume×(mark−entry), Short volume×(entry−mark).
// По залогу (volume ≤ 0): объём = margin × f.Leverage / entry, далее формула по объёму.
func (f Future) UnrealizedPnL(side Side, volume, margin, entry, mark float64) float64 {
	if volume > 0 {
		switch side {
		case Long:
			return volume * (mark - entry)
		case Short:
			return volume * (entry - mark)
		default:
			return 0
		}
	}
	if margin <= 0 || entry <= 0 || f.Leverage <= 0 {
		return 0
	}
	return f.UnrealizedPnL(side, f.VolumeFromMargin(margin, entry), 0, entry, mark)
}

// LiquidationPrice — цена ликвидации для изолированной позиции (USDT-M).
//
// Условие ликвидации: Margin + PnL = MaintenanceMargin = Notional × MMR.
//
// Long:  P_liq = (M - V×E) / (V×(MMR - 1))
// Short: P_liq = (M + V×E) / (V×(1 + MMR))
//
// M — залог (Margin), V — объём, E — средняя цена входа, MMR — maintenance margin rate (например 0.005).
func (f Future) LiquidationPrice(side Side, volume, entry, margin, maintenanceMarginRate float64) float64 {
	if volume <= 0 || margin <= 0 || maintenanceMarginRate <= 0 || maintenanceMarginRate >= 1 {
		return 0
	}
	switch side {
	case Long:
		denom := volume * (maintenanceMarginRate - 1)
		if denom == 0 {
			return 0
		}
		return (margin - volume*entry) / denom
	case Short:
		denom := volume * (1 + maintenanceMarginRate)
		if denom == 0 {
			return 0
		}
		return (margin + volume*entry) / denom
	default:
		return 0
	}
}

// EffectiveLeverage — фактическое плечо: notional / margin.
func (f Future) EffectiveLeverage(volume, price, margin float64) float64 {
	if margin <= 0 {
		return 0
	}
	return Notional(volume, price) / margin
}

// DistanceToLiquidationPercent — запас до ликвидации в % от текущей цены.
//
// Long:  (mark - liq) / mark × 100
// Short: (liq - mark) / mark × 100
func (f Future) DistanceToLiquidationPercent(side Side, mark, liquidation float64) float64 {
	if mark <= 0 || liquidation <= 0 {
		return 0
	}
	switch side {
	case Long:
		return (mark - liquidation) / mark * 100
	case Short:
		return (liquidation - mark) / mark * 100
	default:
		return 0
	}
}

// ResolveAddOnVolume — объём докупки: явный Volume или из Margin × f.Leverage / Price.
func (f Future) ResolveAddOnVolume(add AddOn) float64 {
	if add.Volume > 0 {
		return add.Volume
	}
	return f.VolumeFromMargin(add.Margin, add.Price)
}

// SimulateAddOn — моделирует докупку: новая цена входа, ликвидация, риски.
func (f Future) SimulateAddOn(pos Position, add AddOn, maintenanceMarginRate float64) AddOnResult {
	before := pos
	addVolume := (Future{Leverage: pos.Leverage}).ResolveAddOnVolume(add)

	after := Position{
		Side:       pos.Side,
		Volume:     pos.Volume + addVolume,
		EntryPrice: f.NewAvgEntryPrice(pos.Volume, pos.EntryPrice, addVolume, add.Price),
		Margin:     pos.Margin + add.Margin,
		Leverage:   pos.Leverage,
	}

	liqBefore := f.LiquidationPrice(before.Side, before.Volume, before.EntryPrice, before.Margin, maintenanceMarginRate)
	liqAfter := f.LiquidationPrice(after.Side, after.Volume, after.EntryPrice, after.Margin, maintenanceMarginRate)

	notional := Notional(after.Volume, add.Price)
	pnl := f.UnrealizedPnL(after.Side, after.Volume, 0, after.EntryPrice, add.Price)
	maintMargin := notional * maintenanceMarginRate

	entryDelta := after.EntryPrice - before.EntryPrice
	entryDeltaPct := 0.0
	if before.EntryPrice > 0 {
		entryDeltaPct = entryDelta / before.EntryPrice * 100
	}

	pnlOnMargin := 0.0
	if after.Margin > 0 {
		pnlOnMargin = pnl / after.Margin * 100
	}

	return AddOnResult{
		Before: before,
		After:  after,

		VolumeAdded: addVolume,

		EntryDelta:        entryDelta,
		EntryDeltaPercent: entryDeltaPct,

		LiquidationBefore: liqBefore,
		LiquidationAfter:  liqAfter,
		LiquidationDelta:  liqAfter - liqBefore,

		MarginAdded:       add.Margin,
		NotionalAfter:     notional,
		EffectiveLeverage: f.EffectiveLeverage(after.Volume, add.Price, after.Margin),
		MaintenanceMargin: maintMargin,

		UnrealizedPnLAtPrice:     pnl,
		PnLPercentOnMargin:       pnlOnMargin,
		DistanceToLiquidationPct: f.DistanceToLiquidationPercent(after.Side, add.Price, liqAfter),
		BreakEvenPrice:           after.EntryPrice,
	}
}

// RiskAtPrice — снимок риска позиции при рыночной цене mark.
func (f Future) RiskAtPrice(pos Position, mark, maintenanceMarginRate float64) RiskSnapshot {
	liq := f.LiquidationPrice(pos.Side, pos.Volume, pos.EntryPrice, pos.Margin, maintenanceMarginRate)
	pnl := f.UnrealizedPnL(pos.Side, pos.Volume, 0, pos.EntryPrice, mark)
	notional := Notional(pos.Volume, mark)
	equity := pos.Margin + pnl
	maint := notional * maintenanceMarginRate

	pnlOnMargin := 0.0
	if pos.Margin > 0 {
		pnlOnMargin = pnl / pos.Margin * 100
	}
	marginUsage := 0.0
	if equity > 0 {
		marginUsage = maint / equity * 100
	}

	return RiskSnapshot{
		MarkPrice:                mark,
		EntryPrice:               pos.EntryPrice,
		LiquidationPrice:         liq,
		UnrealizedPnL:            pnl,
		PnLPercentOnMargin:       pnlOnMargin,
		Notional:                 notional,
		EffectiveLeverage:        f.EffectiveLeverage(pos.Volume, mark, pos.Margin),
		MaintenanceMargin:        maint,
		MarginUsagePercent:       marginUsage,
		DistanceToLiquidationPct: f.DistanceToLiquidationPercent(pos.Side, mark, liq),
	}
}

// RoundPrice — округление цены для отображения (8 знаков).
func RoundPrice(v float64) float64 {
	return math.Round(v*1e8) / 1e8
}
