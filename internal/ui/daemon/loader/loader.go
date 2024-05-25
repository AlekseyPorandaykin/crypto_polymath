package loader

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/scheduler"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

// Проскальзывание в секундах, при запросе данные в точное время на стороне сервера данные еще могут не сохраниться.
const slippageSecond = 1

type Loader struct {
	priceService       price.Price
	candlestickService candlestick.Candlestick
	exchangeService    core_exchange.Exchange
	exchangeNames      []string

	symbols []string
}

func NewLoader(
	priceService price.Price,
	candlestickService candlestick.Candlestick,
	exchangeService core_exchange.Exchange,
	exchangeNames,
	symbols []string,
) *Loader {
	return &Loader{
		priceService:       priceService,
		candlestickService: candlestickService,
		exchangeService:    exchangeService,
		exchangeNames:      exchangeNames,
		symbols:            symbols,
	}
}

func (l *Loader) Run(ctx context.Context) error {
	errCh := make(chan error)
	priceDurationLoader, err := time.ParseDuration(viper.GetString("price.duration.loader"))
	if err != nil {
		return errors.Wrap(err, "parse price duration loader")
	}

	for _, exchangeName := range l.exchangeNames {
		go func(exchangeName string) {
			defer system.HandlePanic()
			l.loadExchangePrices(ctx, exchangeName)
			ticker := time.NewTicker(priceDurationLoader)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					ticker.Stop()
					l.loadExchangePrices(ctx, exchangeName)
					ticker.Reset(priceDurationLoader)
				}
			}
		}(exchangeName)
	}
	go func() {
		defer system.HandlePanic()
		for _, symbol := range l.symbols {
			l.loadMinuteCandlesticks(ctx, exchange.BybitExchange, symbol)
		}
	}()
	go func() {
		defer system.HandlePanic()
		for _, symbol := range l.symbols {
			l.loadHourCandlesticks(ctx, exchange.BybitExchange, symbol)
		}
	}()
	go func() {
		defer system.HandlePanic()
		for _, symbol := range l.symbols {
			l.loadDayCandlesticks(ctx, exchange.BybitExchange, symbol)
		}
	}()
	go func() {
		defer system.HandlePanic()
		l.deleteOldRows(ctx, exchange.BybitExchange, viper.GetInt("candlestick.storage.limit"))
	}()
	go func() {
		err := scheduler.ExecuteEveryDay(ctx, func() error {
			if _, err := l.exchangeService.LoadSymbolInfo(ctx, exchange.BybitExchange); err != nil {
				zap.L().Error("load symbol info", zap.Error(err))
			}
			return nil
		})
		if err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		}
	}
}

func (l *Loader) loadExchangePrices(ctx context.Context, exchangeName string) {
	defer durationPricesLoadedHelper(exchangeName)()
	prices, err := l.priceService.LoadPrices(ctx, exchangeName)
	if err != nil {
		zap.L().Error(
			"load prices",
			zap.String("exchangeName", exchangeName),
			zap.Error(err))
	}
	pricesLoaded.WithLabelValues(exchangeName).Add(float64(len(prices)))
}
func (l *Loader) loadMinuteCandlesticks(ctx context.Context, exchangeName, symbol string) {
	for _, interval := range []int{1, 3, 5, 15, 30} {
		minutes := interval
		system.Go(func() {
			//TODO для каждого интервала свое повторение
			maxIteration := 2
			if interval == 1 {
				maxIteration = 5
			}
			_ = scheduler.ExecuteCustomMinuteWithReply(ctx, minutes, slippageSecond/2, 1, maxIteration, func() (bool, error) {
				defer durationCandlestickLoadedHelper(exchangeName, string(domain.MinuteUnit), minutes)()
				candles, _ := l.candlestickService.LoadCandlesticksMinutes(ctx, exchangeName, symbol, minutes)
				return len(candles) > 0, nil
			})
		})
	}
}

func (l *Loader) loadHourCandlesticks(ctx context.Context, exchangeName, symbol string) {
	for _, interval := range []int{1, 2, 4, 6, 12} {
		hours := interval
		system.Go(func() {
			_ = scheduler.ExecuteEveryHour(ctx, hours, slippageSecond, func() error {
				defer durationCandlestickLoadedHelper(exchangeName, string(domain.HourUnit), hours)()
				_, _ = l.candlestickService.LoadCandlesticksHours(ctx, exchangeName, symbol, hours)
				return nil
			})
		})
	}
}

func (l *Loader) loadDayCandlesticks(ctx context.Context, exchangeName, symbol string) {
	system.Go(func() {
		_ = scheduler.ExecuteEveryDay(ctx, func() error {
			defer durationCandlestickLoadedHelper(exchangeName, string(domain.DayUnit), 1)()
			_, _ = l.candlestickService.LoadCandlesticksDay(ctx, exchangeName, symbol)
			return nil
		})
	})
	system.Go(func() {
		_ = scheduler.ExecuteEveryDay(ctx, func() error {
			defer durationCandlestickLoadedHelper(exchangeName, string(domain.WeekUnit), 1)()
			_, _ = l.candlestickService.LoadCandlesticksWeek(ctx, exchangeName, symbol)
			return nil
		})
	})
	system.Go(func() {
		_ = scheduler.ExecuteEveryDay(ctx, func() error {
			defer durationCandlestickLoadedHelper(exchangeName, string(domain.MonthUnit), 1)()
			_, _ = l.candlestickService.LoadCandlesticksMonth(ctx, exchangeName, symbol)
			return nil
		})
	})
}

func (l *Loader) deleteOldRows(ctx context.Context, exchangeName string, oldValueLimit int) {
	defer deleteIndicatorHelper()()
	_ = scheduler.ExecuteEveryDay(ctx, func() error {
		if err := l.priceService.DeleteOldRaws(ctx, exchangeName, time.Now().In(time.UTC).Add(-24*time.Hour)); err != nil {
			zap.L().Error("delete old price", zap.Error(err))
		}
		if err := l.candlestickService.DeleteOldRows(ctx, oldValueLimit); err != nil {
			zap.L().Error("delete old candlestick", zap.Error(err))
		}
		return nil
	})
}
