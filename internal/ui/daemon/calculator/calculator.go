package calculator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/scheduler"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

type Calculator struct {
	indicatorService    indicator.Indicator
	indicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody]
	symbols             []string
}

func NewCalculator(indicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody], indicatorService indicator.Indicator, symbols []string) *Calculator {
	return &Calculator{
		indicatorService:    indicatorService,
		indicatorDispatcher: indicatorDispatcher,
		symbols:             symbols,
	}
}

func (app *Calculator) Run(ctx context.Context) error {
	errCh := make(chan error)
	exchangeNames := []string{exchange.BybitExchange}

	for _, exchangeName := range exchangeNames {
		minutes := viper.GetIntSlice("candlestick.minutes")
		depths := viper.GetIntSlice("candlestick.depths")
		for _, symbol := range app.symbols {
			for _, minute := range minutes {
				for _, depth := range depths {
					go func(exchangeName, symbol string, minute, depth int) {
						defer system.HandlePanic()
						_ = app.execMinMinIndicator(ctx, exchangeName, symbol, minute)

					}(exchangeName, symbol, minute, depth)
				}

			}
			for _, hour := range viper.GetIntSlice("candlestick.hours") {
				go func(exchangeName, symbol string, hour int) {
					defer system.HandlePanic()
					_ = scheduler.ExecuteEveryHour(ctx, 1, 2, func() error {
						app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
							Name: domain.CreateIndicatorEventEvent,
							Body: domain.CreateIndicatorEventBody{
								Exchange: exchangeName,
								Symbol:   symbol,
								Unit:     domain.HourUnit,
								Interval: hour,
							},
						})
						return nil
					})
				}(exchangeName, symbol, hour)
			}
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryDay(ctx, func() error {
					app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
						Name: domain.CreateIndicatorEventEvent,
						Body: domain.CreateIndicatorEventBody{
							Exchange: exchangeName,
							Symbol:   symbol,
							Unit:     domain.DayUnit,
							Interval: 1,
						},
					})
					return nil
				})
			}(exchangeName, symbol)
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryWeek(ctx, func() error {
					app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
						Name: domain.CreateIndicatorEventEvent,
						Body: domain.CreateIndicatorEventBody{
							Exchange: exchangeName,
							Symbol:   symbol,
							Unit:     domain.WeekUnit,
							Interval: 1,
						},
					})
					return nil
				})
			}(exchangeName, symbol)
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryMonth(ctx, func() error {
					app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
						Name: domain.CreateIndicatorEventEvent,
						Body: domain.CreateIndicatorEventBody{
							Exchange: exchangeName,
							Symbol:   symbol,
							Unit:     domain.MonthUnit,
							Interval: 1,
						},
					})
					return nil
				})
			}(exchangeName, symbol)
		}
	}
	system.Go(func() {
		_ = scheduler.ExecuteEveryDay(ctx, func() error {
			return app.deleteOldRows(ctx)
		})
	})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		}
	}
}

func (app *Calculator) execMinMinIndicator(ctx context.Context, exchangeName, symbol string, min int) error {
	ticker := time.NewTicker(time.Second / 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
				Name: domain.CreateIndicatorEventEvent,
				Body: domain.CreateIndicatorEventBody{
					Exchange: exchangeName,
					Symbol:   symbol,
					Unit:     domain.MinuteUnit,
					Interval: min,
				},
			})
		}
	}
}

func (app *Calculator) deleteOldRows(ctx context.Context) error {
	return scheduler.ExecuteEveryDay(ctx, func() error {
		defer deleteIndicatorHelper()()
		if err := app.indicatorService.DeleteOldRows(
			ctx,
			viper.GetInt("indicator.storage.limit"),
		); err != nil {
			zap.L().Error("delete old price", zap.Error(err))
		}
		return nil
	})
}
