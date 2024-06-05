package service

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"go.uber.org/zap"
)

type IndicatorHandler struct {
	indicatorService    indicator.Indicator
	indicatorDispatcher *dispatcher.Dispatcher[domain.Indicator]
	depths              []int

	chans map[string]chan struct{}
}

func NewIndicatorHandler(
	indicatorService indicator.Indicator,
	indicatorDispatcher *dispatcher.Dispatcher[domain.Indicator],
	depths []int,
) *IndicatorHandler {
	return &IndicatorHandler{
		indicatorService:    indicatorService,
		indicatorDispatcher: indicatorDispatcher,
		depths:              depths,
		chans:               make(map[string]chan struct{}),
	}
}

func (i *IndicatorHandler) CalculateByCandle(ctx context.Context, candle domain.Candlestick) {
	key := uniqKey(candle.Exchange, candle.Symbol, candle.Unit, candle.Interval)
	if _, has := i.chans[key]; !has {
		i.chans[key] = make(chan struct{}, 1)
	}
	indicators := make([]domain.Indicator, 0, len(i.depths))
	i.chans[key] <- struct{}{}
	for _, depth := range i.depths {
		data, err := i.indicatorService.CalcIndicatorsByCandlestick(ctx, candle, depth)
		if err != nil {
			zap.L().Error("calculate indicator", zap.Error(err))
			continue
		}
		indicators = append(indicators, data...)
	}
	<-i.chans[key]
	for _, item := range indicators {
		i.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
			Name: domain.CreatedIndicatorEvent,
			Body: item,
		})
	}
}

func (i *IndicatorHandler) Calculate(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int) {
	key := uniqKey(exchange, symbol, unit, interval)
	if _, has := i.chans[key]; !has {
		i.chans[key] = make(chan struct{}, 1)
	}
	indicators := make([]domain.Indicator, 0, len(i.depths))
	i.chans[key] <- struct{}{}
	for _, depth := range i.depths {
		data, err := i.indicatorService.CalcIndicators(ctx, exchange, symbol, unit, interval, depth)
		if err != nil {
			zap.L().Error("calculate indicator", zap.Error(err))
			continue
		}
		indicators = append(indicators, data...)
	}
	<-i.chans[key]
	for _, item := range indicators {
		i.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
			Name: domain.CreatedIndicatorEvent,
			Body: item,
		})
	}
}

func uniqKey(exchange, symbol string, unit domain.Unit, interval int) string {
	return fmt.Sprintf("%s-%s-%s-%d", exchange, symbol, unit, interval)
}
