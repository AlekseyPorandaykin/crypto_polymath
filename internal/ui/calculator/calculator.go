package calculator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/scheduler"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

type Calculator struct {
	candlestickService candlestick.Candlestick
	indicatorService   indicator.Indicator
	symbols            []string
}

func NewCalculator(candlestickService candlestick.Candlestick, indicatorService indicator.Indicator, symbols []string) *Calculator {
	return &Calculator{
		candlestickService: candlestickService,
		indicatorService:   indicatorService,
		symbols:            symbols,
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
						_ = app.execMinMinIndicator(ctx, exchangeName, symbol, minute, depth)

					}(exchangeName, symbol, minute, depth)
				}

			}
			for _, hour := range viper.GetIntSlice("candlestick.hours") {
				go func(exchangeName, symbol string, hour int) {
					defer system.HandlePanic()
					_ = scheduler.ExecuteEveryHour(ctx, 1, 2, func() error {
						return app.calcHourIndicator(ctx, exchangeName, symbol, hour)
					})
				}(exchangeName, symbol, hour)
			}
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryDay(ctx, func() error {
					return app.calcDayIndicator(ctx, exchangeName, symbol)
				})
			}(exchangeName, symbol)
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryWeek(ctx, func() error {
					return app.calcWeekIndicator(ctx, exchangeName, symbol)
				})
			}(exchangeName, symbol)
			go func(exchangeName, symbol string) {
				defer system.HandlePanic()
				_ = scheduler.ExecuteEveryMonth(ctx, func() error {
					return app.calcMonthIndicator(ctx, exchangeName, symbol)
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

func (app *Calculator) execMinMinIndicator(ctx context.Context, exchangeName, symbol string, min, depth int) error {
	ticker := time.NewTicker(time.Second / 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, _ = app.calcMinIndicator(ctx, exchangeName, symbol, min, depth)
		}
	}
}

func (app *Calculator) calcMinIndicator(ctx context.Context, exchangeName, symbol string, min, depth int) (bool, error) {
	count, _ := app.calcIndicator(ctx, exchangeName, symbol, domain.MinuteUnit, min, depth)
	zap.L().Debug("calc min indicator", zap.Int("count", count))
	return count > 0, nil
}

func (app *Calculator) calcHourIndicator(ctx context.Context, exchangeName, symbol string, hour int) error {
	for _, depth := range viper.GetIntSlice("candlestick.depths") {
		_, _ = app.calcIndicator(ctx, exchangeName, symbol, domain.HourUnit, hour, depth)
	}
	return nil
}

func (app *Calculator) calcDayIndicator(ctx context.Context, exchangeName, symbol string) error {
	for _, depth := range viper.GetIntSlice("candlestick.depths") {
		_, _ = app.calcIndicator(ctx, exchangeName, symbol, domain.DayUnit, 1, depth)
	}
	return nil
}

func (app *Calculator) calcWeekIndicator(ctx context.Context, exchangeName, symbol string) error {
	for _, depth := range viper.GetIntSlice("candlestick.depths") {
		_, _ = app.calcIndicator(ctx, exchangeName, symbol, domain.WeekUnit, 1, depth)
	}
	return nil
}
func (app *Calculator) calcMonthIndicator(ctx context.Context, exchangeName, symbol string) error {
	for _, depth := range viper.GetIntSlice("candlestick.depths") {
		_, _ = app.calcIndicator(ctx, exchangeName, symbol, domain.MonthUnit, 1, depth)
	}
	return nil
}
func (app *Calculator) calcIndicator(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval, depth int) (int, error) {
	defer calcIndicatorHelper(exchangeName, string(unit), interval, depth)()
	return app.indicatorService.CalcIndicators(ctx, exchangeName, symbol, unit, interval, depth)
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
