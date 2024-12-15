package loader

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/scheduler"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

// Проскальзывание в секундах, при запросе данные в точное время на стороне сервера данные еще могут не сохраниться.
const slippageSecond = 1

//Loader - это демон, который загружает данные из внешних источников и сохраняет их в базу данных.

type Loader struct {
	priceService       price.Price
	candlestickService candlestick.Candlestick
	exchangeService    core_exchange.Exchange
	indicatorService   indicator.Indicator
	exchangeNames      []string
	candleDispatcher   *dispatcher.Dispatcher[domain.Candlestick]
	service            *application.Service
	candleIndicator    candle_indicator.CandleIndicator

	symbols    []string
	hotSymbols []string

	logger *zap.Logger
}

func NewLoader(
	priceService price.Price,
	candlestickService candlestick.Candlestick,
	exchangeService core_exchange.Exchange,
	indicatorService indicator.Indicator,
	candleDispatcher *dispatcher.Dispatcher[domain.Candlestick],
	service *application.Service,
	candleIndicator candle_indicator.CandleIndicator,
	exchangeNames,
	symbols []string,
	hotSymbols []string,
) *Loader {
	return &Loader{
		priceService:       priceService,
		candlestickService: candlestickService,
		exchangeService:    exchangeService,
		indicatorService:   indicatorService,
		candleDispatcher:   candleDispatcher,
		candleIndicator:    candleIndicator,
		service:            service,
		exchangeNames:      exchangeNames,
		symbols:            util.UniqSlice(append(symbols, hotSymbols...)),
		hotSymbols:         hotSymbols,
		logger:             zap.L(),
	}
}

func (l *Loader) WithLogger(logger *zap.Logger) {
	l.logger = logger
}

func (l *Loader) Run(ctx context.Context) error {
	errCh := make(chan error)
	priceDurationLoader, err := time.ParseDuration(viper.GetString("price.duration.loader"))
	if err != nil {
		return errors.Wrap(err, "parse price duration loader")
	}

	//Загружаем цены
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
	//Загружаем данные по крипто парам
	for _, exchangeName := range []string{exchange.BybitExchange} {
		l.runLoadCandles(ctx, exchangeName)
		go func(exchangeName string) {
			_ = scheduler.ExecuteEveryDay(ctx, func() error {
				defer system.HandlePanic()
				defer ExchangeSymbolLoadedHelper(exchangeName)()
				if _, err := l.exchangeService.LoadSymbolInfo(ctx, exchangeName); err != nil {
					l.logger.Error("load symbol info", zap.Error(err))
				}
				return nil
			})
		}(exchangeName)
	}
	//Загружаем часовые свечи для всех крипто пар
	for _, exchangeName := range []string{exchange.BybitExchange} {
		go func(exchangeName string) {
			defer system.HandlePanic()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					l.loadSymbolFutureCandlesticks(ctx, exchangeName)
				}
			}
		}(exchangeName)
	}

	//Собрать словать для фронта
	go func() {
		_ = scheduler.ExecuteEveryHour(ctx, 1, slippageSecond, func() error {
			defer system.HandlePanic()
			if err := l.service.Collect(ctx); err != nil {
				l.logger.Error("collect", zap.Error(err))
			}
			return nil
		})
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

func (l *Loader) loadSymbolFutureCandlesticks(ctx context.Context, exchangeName string) {
	start := time.Now()
	symbols, errSymbol := l.exchangeService.SymbolInfoByCategory(
		ctx, exchangeName, string(core_exchange.SymbolCategoryFuture),
	)
	if errSymbol != nil {
		l.logger.Error("load symbol info", zap.Error(errSymbol))
		return
	}
	l.logger.Debug("start load future candlesticks", zap.Int("count", len(symbols)))
	if len(symbols) == 0 {
		return
	}
	for _, s := range symbols {
		if !s.IsExist {
			continue
		}
		l.loadFutureCandlesticks(ctx, exchangeName, s.Symbol)
		l.logger.Debug("load future candlesticks", zap.String("symbol", s.Symbol))
	}
	l.logger.Debug("load future candlesticks", zap.String("duration", time.Since(start).String()))
	return
}

func (l *Loader) runLoadCandles(ctx context.Context, exchangeName string) {
	for _, s := range l.symbols {
		symbol := s
		l.loadHourCandlesticks(ctx, exchangeName, symbol)
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
	for _, hs := range l.hotSymbols {
		symbol := hs
		l.loadMinuteCandlesticks(ctx, exchangeName, symbol)
	}
	go func() {
		defer system.HandlePanic()
		l.deleteOldRows(ctx, exchangeName, viper.GetInt("candlestick.storage.limit"))
	}()
}

func (l *Loader) loadExchangePrices(ctx context.Context, exchangeName string) {
	defer durationPricesLoadedHelper(exchangeName)()
	prices, err := l.priceService.LoadPrices(ctx, exchangeName)
	if err != nil {
		l.logger.Error(
			"load prices",
			zap.String("exchangeName", exchangeName),
			zap.Error(err))
	}
	pricesLoaded.WithLabelValues(exchangeName).Add(float64(len(prices)))
}
func (l *Loader) loadMinuteCandlesticks(ctx context.Context, exchangeName, symbol string) {
	for _, interval := range viper.GetIntSlice("candlestick.minutes") {
		minutes := interval
		//TODO для каждого интервала свое повторение
		maxIteration := 2
		if interval == 1 {
			maxIteration = 5
		}
		system.Go(func() {
			_ = scheduler.ExecuteCustomMinuteWithReply(ctx, minutes, slippageSecond/3, 1, maxIteration, func() (bool, error) {
				candles := l.loadCandlesticks(ctx, exchangeName, symbol, domain.MinuteUnit, minutes)
				return len(candles) > 0, nil
			})
		})
	}
}

// Загружаем с задержкой, чтобы не нагрузить апи
func (l *Loader) loadFutureCandlesticks(ctx context.Context, exchangeName, symbol string) {
	d := time.Second * 10
	childCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	l.updateFutureCandlesticks(childCtx, exchangeName, symbol, domain.HourUnit, 1)
	futureCandlestickLoaded.WithLabelValues(exchangeName, "1H").Inc()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTimer(d).C:
			return
		}
	}
}

// loadHourCandlesticks - загружаем все часовые свечи, с разными интервалами.
func (l *Loader) loadHourCandlesticks(ctx context.Context, exchangeName, symbol string) {
	for _, interval := range viper.GetIntSlice("candlestick.hours") {
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
	data := l.saveCandlesticks(ctx, exchangeName, symbol, unit, interval)
	for _, item := range data {
		l.candleDispatcher.Dispatch(dispatcher.Event[domain.Candlestick]{
			Name: domain.CreatedCandlestickEvent,
			Body: item,
		})
	}
	if _, err := l.candleIndicator.CalculateFromCandlesticks(ctx, data); err != nil {
		l.logger.Error("calculate candle indicator", zap.Error(err))
	}
	return data
}

func (l *Loader) saveCandlesticks(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int) []domain.Candlestick {
	defer durationCandlestickLoadedHelper(exchangeName, string(unit), interval)()
	log := l.logger.With(
		zap.String("exchange", exchangeName),
		zap.String("symbol", symbol),
		zap.String("unit", string(unit)),
		zap.Int("interval", interval),
	)
	data, err := l.candlestickService.LoadCandlesticks(ctx, exchangeName, symbol, unit, interval)
	candlestickLoadedTotal.WithLabelValues(exchangeName, string(unit)).Add(float64(len(data)))
	if err != nil {
		errorTotal.WithLabelValues(exchangeName, "load_candlesticks").Inc()
		log.Error("load candlestick", zap.Error(err))
		return nil
	}
	return data
}

func (l *Loader) updateFutureCandlesticks(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int,
) []domain.Candlestick {
	defer durationFutureCandlestickLoadedHelper(exchangeName, fmt.Sprintf("%d%s", interval, string(unit)))()
	log := l.logger.With(
		zap.String("exchange", exchangeName),
		zap.String("symbol", symbol),
		zap.String("unit", string(unit)),
		zap.Int("interval", interval),
	)
	err := l.candlestickService.UpdateCandlesticks(ctx, exchangeName, symbol, unit, interval)
	if err != nil {
		log.Error("update future candlestick", zap.Error(err))
	}
	return nil
}

func (l *Loader) deleteOldRows(ctx context.Context, exchangeName string, oldValueLimit int) {
	defer deleteIndicatorHelper()()
	_ = scheduler.ExecuteEveryDay(ctx, func() error {
		if err := l.priceService.DeleteOldRaws(ctx, exchangeName, time.Now().In(time.UTC).Add(-24*time.Hour)); err != nil {
			l.logger.Error("delete old price", zap.Error(err))
		}
		if err := l.candlestickService.DeleteOldRows(ctx, oldValueLimit); err != nil {
			l.logger.Error("delete old candlestick", zap.Error(err))
		}
		return nil
	})
}
