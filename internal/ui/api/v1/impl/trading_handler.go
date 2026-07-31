package impl

import (
	"errors"
	"net/http"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/labstack/echo/v4"
)

const defaultTradingMMR = 0.005

var errInvalidTradingSide = errors.New("invalid trading side")

func (h *Handler) PostCalculatorTradingAvgEntryPrice(ctx echo.Context) error {
	var req spec.TradingAvgEntryPriceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	return ctx.JSON(http.StatusOK, spec.TradingAvgEntryPriceResponse{
		AvgEntryPrice: float32(trading.Future{}.NewAvgEntryPrice(
			float64(req.EntryVolume),
			float64(req.EntryPrice),
			float64(req.NewVolume),
			float64(req.NewPrice),
		)),
	})
}

func (h *Handler) PostCalculatorTradingAvgEntryPriceBySum(ctx echo.Context) error {
	var req spec.TradingAvgEntryPriceBySumRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	future := trading.Future{Leverage: float64(req.Leverage)}
	return ctx.JSON(http.StatusOK, spec.TradingAvgEntryPriceResponse{
		AvgEntryPrice: float32(future.NewAvgEntryPriceBySum(
			float64(req.EntryVolume),
			float64(req.EntryPrice),
			float64(req.Sum),
			float64(req.NewPrice),
		)),
	})
}

func (h *Handler) PostCalculatorTradingVolumeFromMargin(ctx echo.Context) error {
	var req spec.TradingVolumeFromMarginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	future := trading.Future{Leverage: float64(req.Leverage)}
	return ctx.JSON(http.StatusOK, spec.TradingVolumeResponse{
		Volume: float32(future.VolumeFromMargin(float64(req.Margin), float64(req.Price))),
	})
}

func (h *Handler) PostCalculatorTradingLiquidationPrice(ctx echo.Context) error {
	var req spec.TradingLiquidationPriceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	side, err := tradingSideFromSpec(req.Side)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(http.StatusOK, spec.TradingLiquidationPriceResponse{
		LiquidationPrice: float32(trading.Future{}.LiquidationPrice(
			side,
			float64(req.Volume),
			float64(req.EntryPrice),
			float64(req.Margin),
			tradingMMR(req.MaintenanceMarginRate),
		)),
	})
}

func (h *Handler) PostCalculatorTradingUnrealizedPnl(ctx echo.Context) error {
	var req spec.TradingUnrealizedPnLRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	side, err := tradingSideFromSpec(req.Side)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}
	// Размер позиции задают либо объёмом, либо залогом с плечом — контракт
	// объявляет все три поля необязательными именно поэтому, и неуказанное поле
	// равнозначно нулю.
	volume, margin, leverage := tradingOptional(req.Volume), tradingOptional(req.Margin), tradingOptional(req.Leverage)
	if volume <= 0 && (margin <= 0 || leverage <= 0) {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "volume or margin with leverage is required"})
	}
	future := trading.Future{Leverage: leverage}
	return ctx.JSON(http.StatusOK, spec.TradingUnrealizedPnLResponse{
		UnrealizedPnl: float32(future.UnrealizedPnL(
			side,
			volume,
			margin,
			float64(req.EntryPrice),
			float64(req.MarkPrice),
		)),
	})
}

func (h *Handler) PostCalculatorTradingDistanceToLiquidation(ctx echo.Context) error {
	var req spec.TradingDistanceToLiquidationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	side, err := tradingSideFromSpec(req.Side)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(http.StatusOK, spec.TradingDistanceToLiquidationResponse{
		DistancePercent: float32(trading.Future{}.DistanceToLiquidationPercent(
			side,
			float64(req.MarkPrice),
			float64(req.LiquidationPrice),
		)),
	})
}

func (h *Handler) PostCalculatorTradingSimulateAddOn(ctx echo.Context) error {
	var req spec.TradingSimulateAddOnRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	pos, err := tradingPositionFromSpec(req.Position)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}
	result := trading.Future{}.SimulateAddOn(pos, tradingAddOnFromSpec(req.AddOn), tradingMMR(req.MaintenanceMarginRate))
	return ctx.JSON(http.StatusOK, tradingAddOnResultToSpec(result))
}

func (h *Handler) PostCalculatorTradingRiskAtPrice(ctx echo.Context) error {
	var req spec.TradingRiskAtPriceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	pos, err := tradingPositionFromSpec(req.Position)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}
	snapshot := trading.Future{}.RiskAtPrice(pos, float64(req.MarkPrice), tradingMMR(req.MaintenanceMarginRate))
	return ctx.JSON(http.StatusOK, tradingRiskSnapshotToSpec(snapshot))
}

func (h *Handler) PostCalculatorTradingSpotPnl(ctx echo.Context) error {
	var req spec.TradingSpotPnLRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: "invalid request body"})
	}
	value, percent := trading.Spot{}.PnL(float64(req.Volume), float64(req.EntryPrice), float64(req.MarkPrice))
	return ctx.JSON(http.StatusOK, spec.TradingSpotPnLResponse{
		Value:   float32(value),
		Percent: float32(percent),
	})
}

func tradingSideFromSpec(side spec.TradingSide) (trading.Side, error) {
	switch side {
	case spec.Long:
		return trading.Long, nil
	case spec.Short:
		return trading.Short, nil
	default:
		return 0, errInvalidTradingSide
	}
}

func tradingSideToSpec(side trading.Side) spec.TradingSide {
	if side == trading.Short {
		return spec.Short
	}
	return spec.Long
}

func tradingMMR(rate *float32) float64 {
	if rate == nil {
		return defaultTradingMMR
	}
	return float64(*rate)
}

// tradingOptional читает необязательное число запроса. Ноль вместо nil выбран не
// как «нет значения», а как арифметически нейтральное: расчёты сами проверяют,
// что объём или залог положительны, и отличать пропущенное поле от нулевого им не
// нужно.
func tradingOptional(v *float32) float64 {
	if v == nil {
		return 0
	}
	return float64(*v)
}

func tradingPositionFromSpec(p spec.TradingPosition) (trading.Position, error) {
	side, err := tradingSideFromSpec(p.Side)
	if err != nil {
		return trading.Position{}, err
	}
	return trading.Position{
		Side:       side,
		Volume:     float64(p.Volume),
		EntryPrice: float64(p.EntryPrice),
		Margin:     float64(p.Margin),
		Leverage:   float64(p.Leverage),
	}, nil
}

func tradingPositionToSpec(p trading.Position) spec.TradingPosition {
	return spec.TradingPosition{
		Side:       tradingSideToSpec(p.Side),
		Volume:     float32(p.Volume),
		EntryPrice: float32(p.EntryPrice),
		Margin:     float32(p.Margin),
		Leverage:   float32(p.Leverage),
	}
}

func tradingAddOnFromSpec(a spec.TradingAddOn) trading.AddOn {
	add := trading.AddOn{Price: float64(a.Price)}
	if a.Volume != nil {
		add.Volume = float64(*a.Volume)
	}
	if a.Margin != nil {
		add.Margin = float64(*a.Margin)
	}
	return add
}

func tradingAddOnResultToSpec(r trading.AddOnResult) spec.TradingAddOnResultResponse {
	return spec.TradingAddOnResultResponse{
		Before:                   tradingPositionToSpec(r.Before),
		After:                    tradingPositionToSpec(r.After),
		VolumeAdded:              float32(r.VolumeAdded),
		EntryDelta:               float32(r.EntryDelta),
		EntryDeltaPercent:        float32(r.EntryDeltaPercent),
		LiquidationBefore:        float32(r.LiquidationBefore),
		LiquidationAfter:         float32(r.LiquidationAfter),
		LiquidationDelta:         float32(r.LiquidationDelta),
		MarginAdded:              float32(r.MarginAdded),
		NotionalAfter:            float32(r.NotionalAfter),
		EffectiveLeverage:        float32(r.EffectiveLeverage),
		MaintenanceMargin:        float32(r.MaintenanceMargin),
		UnrealizedPnlAtPrice:     float32(r.UnrealizedPnLAtPrice),
		PnlPercentOnMargin:       float32(r.PnLPercentOnMargin),
		DistanceToLiquidationPct: float32(r.DistanceToLiquidationPct),
		BreakEvenPrice:           float32(r.BreakEvenPrice),
	}
}

func tradingRiskSnapshotToSpec(s trading.RiskSnapshot) spec.TradingRiskSnapshotResponse {
	return spec.TradingRiskSnapshotResponse{
		MarkPrice:                float32(s.MarkPrice),
		EntryPrice:               float32(s.EntryPrice),
		LiquidationPrice:         float32(s.LiquidationPrice),
		UnrealizedPnl:            float32(s.UnrealizedPnL),
		PnlPercentOnMargin:       float32(s.PnLPercentOnMargin),
		Notional:                 float32(s.Notional),
		EffectiveLeverage:        float32(s.EffectiveLeverage),
		MaintenanceMargin:        float32(s.MaintenanceMargin),
		MarginUsagePercent:       float32(s.MarginUsagePercent),
		DistanceToLiquidationPct: float32(s.DistanceToLiquidationPct),
	}
}
