package loader

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
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
	indicatorService   indicator.Indicator
	exchangeNames      []string
	candleDispatcher   *dispatcher.Dispatcher[domain.Candlestick]

	symbols []string
}

func NewLoader(
	priceService price.Price,
	candlestickService candlestick.Candlestick,
	exchangeService core_exchange.Exchange,
	indicatorService indicator.Indicator,
	candleDispatcher *dispatcher.Dispatcher[domain.Candlestick],
	exchangeNames,
	symbols []string,
) *Loader {
	return &Loader{
		priceService:       priceService,
		candlestickService: candlestickService,
		exchangeService:    exchangeService,
		indicatorService:   indicatorService,
		candleDispatcher:   candleDispatcher,
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
	for _, exchangeName := range []string{exchange.BybitExchange} {
		l.runLoadCandles(ctx, exchangeName)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		}
	}
}

func (l *Loader) runLoadCandles(ctx context.Context, exchangeName string) {
	for _, symbol := range l.symbols {
		l.loadHourCandlesticks(ctx, exchangeName, symbol)
		l.loadMinuteCandlesticks(ctx, exchangeName, symbol)
		system.Go(func() {
			_ = scheduler.ExecuteEveryDay(ctx, func() error {
				l.loadCandlesticks(ctx, exchangeName, symbol, domain.DayUnit, 1)
				return nil
			})
		})
		system.Go(func() {
			_ = scheduler.ExecuteEveryDay(ctx, func() error {

				l.loadCandlesticks(ctx, exchangeName, symbol, domain.WeekUnit, 1)
				return nil
			})
		})
		system.Go(func() {
			_ = scheduler.ExecuteEveryDay(ctx, func() error {
				l.loadCandlesticks(ctx, exchangeName, symbol, domain.MonthUnit, 1)
				return nil
			})
		})
	}
	go func() {
		defer system.HandlePanic()
		l.deleteOldRows(ctx, exchangeName, viper.GetInt("candlestick.storage.limit"))
	}()
	go func() {
		_ = scheduler.ExecuteEveryDay(ctx, func() error {
			defer system.HandlePanic()
			defer ExchangeSymbolLoadedHelper(exchangeName)()
			if _, err := l.exchangeService.LoadSymbolInfo(ctx, exchangeName); err != nil {
				zap.L().Error("load symbol info", zap.Error(err))
			}
			return nil
		})
	}()
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
			_ = scheduler.ExecuteCustomMinuteWithReply(ctx, minutes, slippageSecond/3, 1, maxIteration, func() (bool, error) {
				candles := l.loadCandlesticks(ctx, exchangeName, symbol, domain.MinuteUnit, minutes)
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
				l.loadCandlesticks(ctx, exchangeName, symbol, domain.HourUnit, hours)
				return nil
			})
		})
	}
}

func (l *Loader) loadCandlesticks(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int) []domain.Candlestick {
	defer durationCandlestickLoadedHelper(exchangeName, string(unit), interval)()
	log := zap.L().With(
		zap.String("exchange", exchangeName),
		zap.String("symbol", symbol),
		zap.String("unit", string(unit)),
		zap.Int("interval", interval),
	)
	data, err := l.candlestickService.LoadCandlesticks(ctx, exchangeName, symbol, unit, interval)
	candlestickLoadedTotal.WithLabelValues(exchangeName, string(unit)).Add(float64(len(data)))
	if err != nil {
		errorTotal.WithLabelValues(exchangeName, "load_candlesticks").Inc()
		if err != nil {
			log.Error("load candlestick", zap.Error(err))
			return nil
		}
	}
	for _, item := range data {
		l.candleDispatcher.Dispatch(dispatcher.Event[domain.Candlestick]{
			Name: domain.CreatedCandlestickEvent,
			Body: item,
		})
	}
	return data
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
