package impl

import (
	"context"
	"errors"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/view"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/duke-git/lancet/v2/slice"
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
	priceService        price.Price
	candlestickService  candlestick.Candlestick
	indicatorService    indicator.Indicator
	exchangeService     exchange.Exchange
	analysisService     *analysis.Service
	analyticRepository  view.AnalyticInfoRepository
	indicatorRepository view.IndicatorInfoRepository
	serv                *service.Service
	candleIndicator     candle_indicator.CandleIndicator
}

func NewHandler(
	priceService price.Price,
	candlestickService candlestick.Candlestick,
	indicatorService indicator.Indicator,
	exchangeService exchange.Exchange,
	analysisService *analysis.Service,
	analyticRepository view.AnalyticInfoRepository,
	indicatorRepository view.IndicatorInfoRepository,
	serv *service.Service,
	candleIndicator candle_indicator.CandleIndicator,
) *Handler {
	return &Handler{
		priceService:        priceService,
		candlestickService:  candlestickService,
		indicatorService:    indicatorService,
		exchangeService:     exchangeService,
		analysisService:     analysisService,
		analyticRepository:  analyticRepository,
		indicatorRepository: indicatorRepository,
		serv:                serv,
		candleIndicator:     candleIndicator,
	}
}

func (h *Handler) GetAnalysisExchangeSymbolUnitIntervalNameIndicatorDepthDepth(
	ctx echo.Context,
	exchange string,
	symbol string,
	unit spec.GetAnalysisExchangeSymbolUnitIntervalNameIndicatorDepthDepthParamsUnit,
	interval int,
	name spec.GetAnalysisExchangeSymbolUnitIntervalNameIndicatorDepthDepthParamsName,
	indicatorDepth spec.GetAnalysisExchangeSymbolUnitIntervalNameIndicatorDepthDepthParamsIndicatorDepth,
	depth spec.GetAnalysisExchangeSymbolUnitIntervalNameIndicatorDepthDepthParamsDepth,
) error {
	data, err := h.analysisService.LastAnalytics(
		ctx.Request().Context(),
		exchange,
		symbol,
		domain.Unit(unit),
		interval,
		string(name),
		int(indicatorDepth),
		int(depth),
	)
	if err != nil {
		return errorResponse(ctx, err)
	}
	resp := make(spec.AnalysisResponse, 0, len(data))
	for _, item := range data {
		resp = append(resp, spec.AnalysisItem{
			Datetime: item.Datetime,
			Value:    float32(item.Value),
		})
	}
	slice.SortBy(resp, func(a, b spec.AnalysisItem) bool {
		return a.Datetime.After(b.Datetime)
	})
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GetExchangeExchangeSymbol(ctx echo.Context, exchange string, symbol string) error {
	res, err := h.exchangeService.SymbolInfo(ctx.Request().Context(), exchange, symbol)
	if err != nil {
		return errorResponse(ctx, err)
	}
	if res == nil {
		return ctx.JSON(http.StatusNotFound, nil)
	}
	return ctx.JSON(http.StatusOK, spec.SymbolInfoResponse{
		Exchange:   res.Exchange,
		Symbol:     res.Symbol,
		BaseAsset:  res.BaseAsset,
		QuoteAsset: res.QuoteAsset,
		IsExist:    res.IsExist,
	})
}

func (h *Handler) GetSymbolsExchangeCategory(
	ctx echo.Context, exchange string, category spec.GetSymbolsExchangeCategoryParamsCategory,
) error {
	res, err := h.exchangeService.SymbolInfoByCategory(ctx.Request().Context(), exchange, string(category))
	if err != nil {
		return errorResponse(ctx, err)
	}
	result := make([]spec.SymbolInfoResponse, 0, len(res))
	now := time.Now().In(time.UTC)
	for _, item := range res {
		si := spec.SymbolInfoResponse{
			Exchange:    item.Exchange,
			Symbol:      item.Symbol,
			BaseAsset:   item.BaseAsset,
			QuoteAsset:  item.QuoteAsset,
			IsExist:     item.IsExist,
			FundingRate: item.FundingRate,
		}
		if item.NextFundingTime != nil && item.NextFundingTime.After(now) {
			si.NextFundingTime = item.NextFundingTime
			si.CountdownFundingTimeSeconds = float32(item.CountdownFundingTime().Seconds())
			si.CountdownFundingTime = item.CountdownFundingTime().String()
		}
		result = append(result, si)
	}
	return ctx.JSON(http.StatusOK, result)
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
	_ = slice.SortByField(result, "Symbol", "ASC")
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
	_ = slice.SortByField(result, "Exchange", "ASC")
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
	data, err := h.candlestickService.Candlesticks(
		ctx.Request().Context(), exchange, symbol, domain.Unit(unit), intervalVal, 100,
	)
	if err != nil {
		return errorResponse(ctx, err)
	}
	response := spec.CandlesticksResponse{}
	for _, item := range data {
		response = append(response, spec.CandlestickItem{
			ClosePrice:        float32(item.ClosePrice),
			HighPrice:         float32(item.HighPrice),
			LowPrice:          float32(item.LowPrice),
			OpenPrice:         float32(item.OpenPrice),
			StartTime:         item.StartTime,
			Volume:            float32(item.Volume),
			CloseLocation:     float32(item.CloseLocation()),
			OpenLocation:      float32(item.OpenLocation()),
			Direction:         spec.CandlestickItemDirection(item.Direction()),
			SizeBodyInPercent: float32(item.SizeBodyInPercent()),
		})
	}
	return ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetServer(ctx echo.Context) error {
	data, err := h.serv.Dictionary(ctx.Request().Context())
	if err != nil {
		return errorResponse(ctx, err)
	}
	unitIntervals := make([]spec.UnitIntervals, 0, 10)
	analysisData := make([]spec.AnalysisInfo, 0, 10)
	indicatorData := make([]spec.IndicatorInfo, 0, 10)

	for _, analyticInfoItem := range data.Analysis {
		analysisData = append(analysisData, spec.AnalysisInfo{
			Name:           analyticInfoItem.Name,
			Description:    analyticInfoItem.Description,
			Depth:          analyticInfoItem.Depth,
			IndicatorDepth: analyticInfoItem.IndicatorDepth,
		})
	}
	for _, indicatorInfoItem := range data.Indicators {
		indicatorData = append(indicatorData, spec.IndicatorInfo{
			Name:        indicatorInfoItem.Name,
			Description: indicatorInfoItem.Description,
			Depth:       indicatorInfoItem.Depth,
		})
	}
	for _, intervalItem := range data.Intervals {
		unitIntervals = append(unitIntervals, spec.UnitIntervals{
			Unit:   intervalItem.Unit,
			Values: intervalItem.Values,
		})
	}

	return ctx.JSON(http.StatusOK, spec.ServerInfoResponse{
		Analysis:       analysisData,
		Depths:         data.Depths,
		IndicatorDepth: data.IndicatorDepth,
		Exchanges:      data.Exchanges,
		Indicators:     indicatorData,
		Intervals:      unitIntervals,
		Symbols:        data.Symbols,
		Time:           time.Now().In(time.UTC),
		Units:          data.Units,
	})
}

func (h *Handler) GetCandleIndicatorExchangeSymbolUnitIntervalName(ctx echo.Context, exchange string, symbol string, unit spec.GetCandleIndicatorExchangeSymbolUnitIntervalNameParamsUnit, interval int, name spec.GetCandleIndicatorExchangeSymbolUnitIntervalNameParamsName) error {
	data, err := h.candleIndicator.Indicators(ctx.Request().Context(), string(name), exchange, symbol, domain.Unit(unit), interval)
	if err != nil {
		return errorResponse(ctx, err)
	}
	response := make(spec.CandleIndicatorResponse, 0, 100)
	for _, item := range data {
		response = append(response, spec.CandleIndicatorItem{
			ClosePrice:        float32(item.ClosePrice),
			HighPrice:         float32(item.HighPrice),
			LowPrice:          float32(item.LowPrice),
			OpenLocation:      float32(item.OpenLocation()),
			OpenPrice:         float32(item.OpenPrice),
			SizeBody:          float32(item.SizeBody()),
			SizeBodyInPercent: float32(item.SizeBodyInPercent()),
			StartTime:         item.StartTime,
			CloseLocation:     float32(item.CloseLocation()),
			Direction:         spec.CandleIndicatorItemDirection(item.Direction()),
		})
	}

	slice.SortBy(response, func(a, b spec.CandleIndicatorItem) bool {
		return a.StartTime.After(b.StartTime)
	})
	return ctx.JSON(http.StatusOK, response)
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
	data, err := h.indicatorService.CalculateLastSequence(
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
	slice.SortBy(indicators, func(a, b spec.IndicatorItem) bool {
		return a.Datetime.After(b.Datetime)
	})
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
	if errors.Is(err, context.Canceled) {
		return nil
	}
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
