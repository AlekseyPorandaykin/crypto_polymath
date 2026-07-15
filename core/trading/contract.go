package trading

// Side — направление позиции на фьючерсах.
type Side int

const (
	Long  Side = 1
	Short Side = -1
)

// Position — изолированная маржинальная позиция (USDT-M).
type Position struct {
	Side       Side
	Volume     float64 // объём в базовом активе (BTC, ETH, …)
	EntryPrice float64 // средняя цена входа
	Margin     float64 // залог в котируемой валюте (USDT)
	Leverage   float64 // плечо для перевода суммы докупки в объём
}

// Future — плечевой (фьючерсный/маржинальный) расчётчик: методы, которым для расчёта
// нужно плечо (Leverage), например перевод суммы залога в объём позиции.
type Future struct {
	Leverage float64
}

// AddOn — докупка / добавление к позиции.
// Задаётся либо объёмом (Volume > 0), либо суммой залога (Margin > 0, Volume = 0).
type AddOn struct {
	Price  float64 // цена докупки
	Volume float64 // доп. объём; если 0 — считается из Margin × Leverage / Price
	Margin float64 // доп. залог (USDT), добавляется к Margin позиции
}

// AddOnResult — снимок до/после докупки и производные метрики.
type AddOnResult struct {
	Before Position
	After  Position

	VolumeAdded float64

	EntryDelta        float64 // After.Entry - Before.Entry
	EntryDeltaPercent float64 // в % от цены входа до докупки

	LiquidationBefore float64
	LiquidationAfter  float64
	LiquidationDelta  float64 // After - Before (для лонга: отриц. = ликвидация ниже)

	MarginAdded       float64
	NotionalAfter     float64 // Volume × Price на момент докупки
	EffectiveLeverage float64 // Notional / Margin после докупки
	MaintenanceMargin float64 // Notional × MMR после докупки

	UnrealizedPnLAtPrice     float64 // PnL по цене докупки
	PnLPercentOnMargin       float64 // PnL / Margin × 100
	DistanceToLiquidationPct float64 // запас до ликвидации в % от цены
	BreakEvenPrice           float64 // цена безубытка (≈ Entry после усреднения)
}

// VolatilityMeasure — какая метрика волатильности используется для расчёта SL.
type VolatilityMeasure int

const (
	// VolatilityRange — (high−low)/close × 100 на HA-свече; лучший результат в бэктесте.
	VolatilityRange VolatilityMeasure = iota
	// VolatilityATR — средний диапазон свечей за lookback.
	VolatilityATR
	// VolatilityCombined — 0.5×ATR + 0.3×std(returns) + 0.2×range.
	VolatilityCombined
)

// StopLossCoefficients — коэффициенты динамического SL от волатильности.
// Значения по умолчанию получены бэктестом на storage/data (499 символов, ~100 часов).
type StopLossCoefficients struct {
	KSL           float64 // множитель начального SL: sl_price_% = KSL × vol
	KTrail        float64 // множитель активации трейлинга
	SLFloorPct    float64 // минимум SL в % от цены (для низковолатильного рынка)
	TrailFloorPct float64 // минимум порога трейлинга в % от цены
	DojiBodyPct   float64 // порог доджи HA: body/range × 100 ≤ DojiBodyPct
	SLCapPct      float64 // максимум SL в % от цены (0 = без ограничения)
	TrailCapPct   float64 // максимум порога трейлинга в % от цены
}

// VolatilitySnapshot — волатильность на свече подтверждения сигнала.
type VolatilitySnapshot struct {
	RangePct float64 // (HA_high − HA_low) / close × 100
	ATRPct   float64 // средний (high−low)/close за lookback
	MktVol   float64 // комбинированная мера
}

// DynamicStopLoss — расчётные уровни SL и активации трейлинга.
type DynamicStopLoss struct {
	VolatilityUsed         float64 // значение vol, использованное в формуле
	SLPricePct             float64 // дистанция SL от входа в % цены
	TrailActivatePricePct  float64 // движение цены для активации трейлинга, %
	InitialStopPrice       float64 // цена начального SL
	TrailActivatePrice     float64 // цена активации трейлинга (в сторону профита)
	InitialSLMarginPct     float64 // отрицательный PnL на марже при срабатывании SL
	TrailActivateMarginPct float64 // положительный PnL на марже для включения трейлинга
}

// TrailingStopUpdate — результат обновления трейлинг-SL за HA-свечу.
type TrailingStopUpdate struct {
	StopPrice      float64
	TrailActivated bool
}

// RiskSnapshot — риск позиции при заданной рыночной цене.
type RiskSnapshot struct {
	MarkPrice                float64
	EntryPrice               float64
	LiquidationPrice         float64
	UnrealizedPnL            float64
	PnLPercentOnMargin       float64
	Notional                 float64
	EffectiveLeverage        float64
	MaintenanceMargin        float64
	MarginUsagePercent       float64 // MaintenanceMargin / (Margin + PnL) × 100
	DistanceToLiquidationPct float64
}
