package impl

import (
	"errors"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

var _ spec.ServerInterface = (*Handler)(nil)

var (
	errNotFoundUnit      = errors.New("not found unit")
	errIncorrectInterval = errors.New("incorrect interval")
	errIncorrectDepth    = errors.New("incorrect depth")
)

type Handler struct {
	priceService       price.Price
	candlestickService candlestick.Candlestick
	indicatorService   indicator.Indicator
}

func NewHandler(priceService price.Price, candlestickService candlestick.Candlestick, indicatorService indicator.Indicator) *Handler {
	return &Handler{priceService: priceService, candlestickService: candlestickService, indicatorService: indicatorService}
}

func (h *Handler) GetPriceExchangeSymbol(ctx echo.Context, exchange string, symbol string) error {
	lastPrice, err := h.priceService.LastPrice(ctx.Request().Context(), exchange, symbol)
	if err != nil {
		return errorResponse(ctx, err)
	}
	if lastPrice == nil {
		return ctx.JSON(http.StatusNotFound, nil)
	}
	return ctx.JSON(http.StatusOK, spec.PriceResponse{
		Exchange: lastPrice.Exchange,
		Symbol:   lastPrice.Symbol,
		Value:    float32(lastPrice.Value),
	})
}

func (h *Handler) GetPricesExchangeExchange(ctx echo.Context, exchange string) error {
	prices, err := h.priceService.LastPricesByExchange(ctx.Request().Context(), exchange)
	if err != nil {
		return errorResponse(ctx, err)
	}
	if len(prices) == 0 {
		return ctx.JSON(http.StatusNotFound, nil)
	}
	result := make(spec.PricesResponse, 0, len(prices))
	for _, item := range prices {
		result = append(result, spec.PriceResponse{
			Exchange: item.Exchange,
			Symbol:   item.Symbol,
			Value:    float32(item.Value),
		})
	}
	return ctx.JSON(http.StatusOK, result)
}

func (h *Handler) GetPricesSymbolSymbol(ctx echo.Context, symbol string) error {
	prices, err := h.priceService.LastPricesBySymbol(ctx.Request().Context(), symbol)
	if err != nil {
		return errorResponse(ctx, err)
	}
	if len(prices) == 0 {
		return ctx.JSON(http.StatusNotFound, nil)
	}
	result := make(spec.PricesResponse, 0, len(prices))
	for _, item := range prices {
		result = append(result, spec.PriceResponse{
			Exchange: item.Exchange,
			Symbol:   item.Symbol,
			Value:    float32(item.Value),
		})
	}
	return ctx.JSON(http.StatusOK, result)
}

func (h *Handler) GetCandlestickExchangeSymbolUnitInterval(
	ctx echo.Context,
	exchange string,
	symbol string,
	unit spec.GetCandlestickExchangeSymbolUnitIntervalParamsUnit,
	interval string,
) error {
	intervalVal, err := strconv.Atoi(interval)
	if err != nil {
		return errorResponse(ctx, errIncorrectInterval)
	}
	if err := checkUnitInterval(string(unit), intervalVal); err != nil {
		return errorResponse(ctx, err)
	}
	data, err := h.candlestickService.Candlestick(
		ctx.Request().Context(), exchange, symbol, domain.Unit(unit), intervalVal, 100,
	)
	if err != nil {
		return errorResponse(ctx, err)
	}
	response := spec.CandlesticksResponse{}
	for _, item := range data {
		_ = item
		response = append(response, spec.CandlestickItem{
			ClosePrice: float32(item.ClosePrice),
			HighPrice:  float32(item.HighPrice),
			LowPrice:   float32(item.LowPrice),
			OpenPrice:  float32(item.OpenPrice),
			StartTime:  item.StartTime,
			Volume:     float32(item.Volume),
		})
	}
	return ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetServer(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, spec.ServerInfoResponse{Time: time.Now().In(time.UTC)})
}

func (h *Handler) GetIndicatorExchangeSymbolUnitIntervalNameDepth(
	ctx echo.Context,
	exchange string,
	symbol string,
	unit spec.GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnit,
	interval int,
	name spec.GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsName,
	depth spec.GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsDepth,
) error {
	if err := checkUnitInterval(string(unit), interval); err != nil {
		return errorResponse(ctx, err)
	}
	if err := checkDepth(int(depth)); err != nil {
		return errorResponse(ctx, err)
	}
	data, err := h.indicatorService.Indicators(
		ctx.Request().Context(), exchange, symbol, domain.Unit(unit), interval, string(name), int(depth), 100,
	)
	if err != nil {
		return errorResponse(ctx, err)
	}
	indicators := make(spec.IndicatorResponse, 0, len(data))
	for _, item := range data {
		indicators = append(indicators, spec.IndicatorItem{
			Datetime: item.Datetime,
			Value:    float32(item.Value),
		})
	}
	return ctx.JSON(http.StatusOK, indicators)
}

func checkDepth(depth int) error {
	for _, item := range viper.GetIntSlice("candlestick.depths") {
		if item == depth {
			return nil
		}
	}
	return errIncorrectDepth
}

func checkUnitInterval(unit string, interval int) error {
	switch domain.Unit(unit) {
	case domain.MinuteUnit:
		for _, item := range viper.GetIntSlice("candlestick.minutes") {
			if interval == item {
				return nil
			}
		}
		return errIncorrectInterval
	case domain.HourUnit:
		for _, item := range viper.GetIntSlice("candlestick.hours") {
			if interval == item {
				return nil
			}
		}
		return errIncorrectInterval
	case domain.DayUnit, domain.WeekUnit, domain.MonthUnit:
		if interval == 1 {
			return nil
		}
		return errIncorrectInterval
	default:
		return errNotFoundUnit
	}
}

func errorResponse(ctx echo.Context, err error) error {
	if errors.Is(err, errNotFoundUnit) {
		return ctx.JSON(http.StatusInternalServerError, spec.ErrorResponse{Message: "not found unit"})
	}
	if errors.Is(err, errIncorrectInterval) {
		return ctx.JSON(http.StatusInternalServerError, spec.ErrorResponse{Message: "incorrect interval"})
	}
	if errors.Is(err, errIncorrectDepth) {
		return ctx.JSON(http.StatusInternalServerError, spec.ErrorResponse{Message: "incorrect depth"})
	}
	zap.L().Error("error response", zap.Error(err))
	return ctx.JSON(http.StatusInternalServerError, spec.ErrorResponse{Message: "internal error"})
}
