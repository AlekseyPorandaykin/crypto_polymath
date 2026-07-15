package trading

// Spot — расчёты для спотовой позиции: без плеча, без ликвидации и margin call.
type Spot struct{}

// PnL — прибыль/убыток спотовой позиции в валюте котировки и в процентах от цены входа.
// value   = volume × (mark - entry)
// percent = (mark - entry) / entry × 100
func (s Spot) PnL(volume, entry, mark float64) (value, percent float64) {
	value = volume * (mark - entry)
	if entry > 0 {
		percent = (mark - entry) / entry * 100
	}
	return value, percent
}
