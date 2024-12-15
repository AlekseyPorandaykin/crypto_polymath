package calculator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/scheduler"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Calculator -	компонент, который отвечает за расчет индикаторов и аналитики.
// Запускается по в виде демона и выполняет расчеты по расписанию.
type Calculator struct {
	indicatorService    indicator.Indicator
	analysisService     *analysis.Service
	indicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody]
	symbols             []string
	logger              *zap.Logger
}

func NewCalculator(
	indicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody],
	indicatorService indicator.Indicator,
	analysisService *analysis.Service,
	symbols []string,
) *Calculator {
	return &Calculator{
		indicatorService:    indicatorService,
		analysisService:     analysisService,
		indicatorDispatcher: indicatorDispatcher,
		symbols:             symbols,
		logger:              zap.L(),
	}
}

func (app *Calculator) WithLogger(logger *zap.Logger) {
	app.logger = logger
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
						_ = scheduler.ExecuteCustomMinute(ctx, minute, 1, func() error {
							app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.CreateIndicatorEventBody]{
								Name: domain.CreateIndicatorEventEvent,
								Body: domain.CreateIndicatorEventBody{
									Exchange: exchangeName,
									Symbol:   symbol,
									Unit:     domain.MinuteUnit,
									Interval: minute,
								},
							})
							return nil
						})
						//_ = app.execMinMinIndicator(ctx, exchangeName, symbol, minute)

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
		_ = scheduler.ExecuteEveryHour(ctx, 1, 1, func() error {
			return app.deleteOlIndicators(ctx)
		})
	})
	system.Go(func() {
		_ = scheduler.ExecuteEveryHour(ctx, 1, 1, func() error {
			return app.deleteOldAnalysis(ctx)
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

func (app *Calculator) deleteOlIndicators(ctx context.Context) error {
	return scheduler.ExecuteEveryDay(ctx, func() error {
		defer deleteIndicatorHelper()()
		if err := app.indicatorService.DeleteOldRows(
			ctx,
			viper.GetInt("indicator.storage.limit"),
		); err != nil {
			app.logger.Error("delete old indicators", zap.Error(err))
		}
		return nil
	})
}

func (app *Calculator) deleteOldAnalysis(ctx context.Context) error {
	return scheduler.ExecuteEveryDay(ctx, func() error {
		defer deleteAnalysisHelper()()
		if err := app.analysisService.DeleteOldRows(
			ctx,
			viper.GetInt("analysis.storage.limit"),
		); err != nil {
			app.logger.Error("delete old analysis", zap.Error(err))
		}
		return nil
	})
}
