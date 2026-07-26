package loader

import (
	"context"
	"fmt"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-kit/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-kit/pkg/scheduler"
	"github.com/AlekseyPorandaykin/go-kit/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	service            *service.Service

	loadedCandleDispatcher dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody]
	priceDispatcher        dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody]
	candleDispatcher       dispatcher.Dispatcher[domain.Candlestick]
	symbols                []string
	hotSymbols             []string

	logger *zap.Logger
}

func NewLoader(
	priceService price.Price,
	candlestickService candlestick.Candlestick,
	exchangeService core_exchange.Exchange,
	indicatorService indicator.Indicator,
	service *service.Service,
	loadedCandleDispatcher dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody],
	priceDispatcher dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody],
	candleDispatcher dispatcher.Dispatcher[domain.Candlestick],
	exchangeNames,
	symbols []string,
	hotSymbols []string,
) *Loader {
	return &Loader{
		priceService:           priceService,
		candlestickService:     candlestickService,
		exchangeService:        exchangeService,
		indicatorService:       indicatorService,
		service:                service,
		exchangeNames:          exchangeNames,
		symbols:                util.UniqSlice(append(symbols, hotSymbols...)),
		hotSymbols:             hotSymbols,
		logger:                 zap.NewNop(),
		loadedCandleDispatcher: loadedCandleDispatcher,
		candleDispatcher:       candleDispatcher,
		priceDispatcher:        priceDispatcher,
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
			logger := l.logger.With(zap.String("exchange ", exchangeName))
			ticker := time.NewTicker(priceDurationLoader)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					ticker.Stop()
					start := time.Now()
					l.loadExchangePrices(ctx, exchangeName)
					logger.Info("loaded prices", zap.String("duration", time.Since(start).String()))
					ticker.Reset(priceDurationLoader)
				}
			}
		}(exchangeName)

		//Удаляем старые записи
		go func() {
			defer system.HandlePanic()
			start := time.Now()
			l.runDeleteOldRows(ctx, exchangeName, viper.GetInt("candlestick.storage.limit"))
			l.logger.Info("Delete old rows", zap.String("duration", time.Since(start).String()))
		}()
		//Загружаем свечи
		l.runLoadCandles(ctx, exchangeName)
		//Загружаем данные по крипто парам
		go func(exchangeName string) {
			_ = scheduler.ExecuteCustomMinute(ctx, 5, 1, func() error {
				defer system.HandlePanic()
				defer ExchangeSymbolLoadedHelper(exchangeName)()
				logger := l.logger.With(zap.String("exchange ", exchangeName))
				start := time.Now()
				data, errSy := l.exchangeService.LoadSymbolInfo(ctx, exchangeName)
				if errSy != nil {
					logger.Error("load symbol info", zap.Error(errSy))
				}
				logger.Info(
					"load symbol infos",
					zap.String("duration", time.Since(start).String()),
					zap.Int("count", len(data)),
				)
				return nil
			})
		}(exchangeName)

		//Загружаем часовые свечи для всех крипто пар
		go func(exchangeName string) {
			defer system.HandlePanic()
			//Загружаем свечи по всем future, можем подождать 30 секунд, чтобы все данные существовали.
			_ = scheduler.ExecuteEveryHour(ctx, 1, 30, func() error {
				start := time.Now()
				l.loadSymbolFutureCandlesticks(ctx, exchangeName)
				l.logger.Info("Load symbol future candlesticks", zap.String("duration", time.Since(start).String()))
				return nil
			})
		}(exchangeName)
	}

	//Собрать словарь для фронта
	go func() {
		_ = scheduler.ExecuteEveryHour(ctx, 1, slippageSecond, func() error {
			defer system.HandlePanic()
			start := time.Now()
			if err := l.service.Collect(ctx); err != nil {
				l.logger.Error("collect", zap.Error(err))
			}
			l.logger.Info("Collect dictionary", zap.String("duration", time.Since(start).String()))
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

// symbolsWithRetry - получаем список символов с ретраем, так как данные могут не сразу появиться.
func (l *Loader) symbolsWithRetry(ctx context.Context, exchangeName string) ([]string, error) {
	retry := 3
	symbols := make([]string, 0, 1000)
	for i := 0; i < retry; i++ {
		resp, err := l.exchangeService.SymbolInfoByCategory(
			ctx, exchangeName, string(core_exchange.SymbolCategoryFuture),
		)
		if err != nil {
			return nil, err
		}
		if len(resp) == 0 {
			durations := time.Second * 30 * time.Duration(i+1)
			time.Sleep(durations)
			continue
		}
		mainSymbols := make([]string, 0, len(resp))
		additionalSymbols := make([]string, 0, len(resp))
		for _, s := range resp {
			if !s.IsExist {
				continue
			}
			if s.QuoteAsset == domain.MainQuoteAsset {
				mainSymbols = append(mainSymbols, s.Symbol)
				continue
			}
			additionalSymbols = append(additionalSymbols, s.Symbol)
		}

		symbols = append(symbols, mainSymbols...)
		symbols = append(symbols, additionalSymbols...)
		break
	}
	return symbols, nil
}

func (l *Loader) loadSymbolFutureCandlesticks(ctx context.Context, exchangeName string) {
	start := time.Now()
	symbols, errSymbol := l.symbolsWithRetry(ctx, exchangeName)
	if errSymbol != nil {
		l.logger.Error("load symbol info", zap.Error(errSymbol))
		return
	}
	l.logger.Debug("start load future candlesticks", zap.Int("count", len(symbols)))
	if len(symbols) == 0 {
		return
	}
	for _, s := range symbols {
		l.loadFutureCandlesticks(ctx, exchangeName, s)
	}
	l.logger.Debug(
		"load future candlesticks",
		zap.String("exchange", exchangeName),
		zap.String("duration", time.Since(start).String()),
	)
	futureCandlestickLoadedTotalDuration.WithLabelValues(exchangeName).Add(time.Since(start).Seconds())
	return
}

func (l *Loader) runLoadCandles(ctx context.Context, exchangeName string) {
	for _, s := range l.symbols {
		symbol := s
		for _, interval := range viper.GetIntSlice("candlestick.hours") {
			hours := interval
			system.Go(func() {
				_ = scheduler.ExecuteEveryHour(ctx, hours, slippageSecond, func() error {
					l.loadCandlesticks(ctx, exchangeName, symbol, domain.HourUnit, hours)
					return nil
				})
			})
		}
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
}

func (l *Loader) loadExchangePrices(ctx context.Context, exchangeName string) {
	defer durationPricesLoadedHelper(exchangeName)()
	startLoad := time.Now()
	prices, err := l.priceService.LoadPrices(ctx, exchangeName)
	if err != nil {
		l.logger.Error(
			"load prices",
			zap.String("exchangeName", exchangeName),
			zap.Error(err))
	}

	l.priceDispatcher.Dispatch(dispatcher.Event[domain.LoadedPricesByExchangeActionBody]{
		Body: domain.LoadedPricesByExchangeActionBody{
			Exchange:  exchangeName,
			CreatedAt: time.Now(),
			Duration:  time.Since(startLoad),
		},
	})
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

func (l *Loader) loadFutureCandlesticks(ctx context.Context, exchangeName, symbol string) {
	startLoadSymbol := time.Now()
	// Можем упереться на стороне клиента на 11 минут, если запросов будет много.
	childCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	data := l.updateFutureCandlesticks(childCtx, exchangeName, symbol, domain.HourUnit, 1)
	futureCandlestickLoaded.WithLabelValues(exchangeName, "1H").Inc()
	l.logger.Debug("load future candlesticks",
		zap.String("symbol", symbol),
		zap.String("duration", time.Since(startLoadSymbol).String()),
	)
	l.loadedCandleDispatcher.Dispatch(dispatcher.Event[domain.LoadedCandlesticksActionBody]{
		Name: domain.LoadedCandlesticksForSymbolAction,
		Body: domain.LoadedCandlesticksActionBody{
			Exchange:  exchangeName,
			Symbol:    symbol,
			Unit:      domain.HourUnit,
			Interval:  1,
			CreatedAt: time.Now(),
			Duration:  time.Since(startLoadSymbol),
		},
	})
	for i := range data {
		l.candleDispatcher.Dispatch(dispatcher.Event[domain.Candlestick]{
			Name: domain.CreatedCandlestickEvent,
			Body: data[i],
		})
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
	defer durationCandlestickLoadedHelper(exchangeName, string(unit), interval)()
	log := l.logger.With(
		zap.String("exchange", exchangeName),
		zap.String("symbol", symbol),
		zap.String("unit", string(unit)),
		zap.Int("interval", interval),
	)
	startLoadSymbol := time.Now()
	data, err := l.candlestickService.LoadCandlesticks(ctx, exchangeName, symbol, unit, interval)
	log = log.With(zap.String("duration", time.Since(startLoadSymbol).String()))
	candlestickLoadedTotal.WithLabelValues(exchangeName, string(unit)).Add(float64(len(data)))
	if err != nil {
		errorTotal.WithLabelValues(exchangeName, "load_candlesticks").Inc()
		log.Error("load candlestick", zap.Error(err))
		return nil
	}
	log.Debug("load candlesticks")
	l.loadedCandleDispatcher.Dispatch(dispatcher.Event[domain.LoadedCandlesticksActionBody]{
		Name: domain.LoadedCandlesticksForSymbolAction,
		Body: domain.LoadedCandlesticksActionBody{
			Exchange:  exchangeName,
			Symbol:    symbol,
			Unit:      unit,
			Interval:  interval,
			CreatedAt: time.Now(),
			Duration:  time.Since(startLoadSymbol),
		},
	})
	for i := range data {
		if data[i].Unit != domain.MinuteUnit && data[i].Interval == 1 {
			l.candleDispatcher.Dispatch(dispatcher.Event[domain.Candlestick]{
				Name: domain.CreatedCandlestickEvent,
				Body: data[i],
			})
		}
	}
	return data
}

func (l *Loader) updateFutureCandlesticks(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int,
) []domain.Candlestick {
	defer durationFutureCandlestickLoadedHelper(exchangeName, fmt.Sprintf("%d%s", interval, string(unit)))()
	data, err := l.candlestickService.UpdateCandlesticks(ctx, exchangeName, symbol, unit, interval)
	if err != nil {
		l.logger.With(
			zap.String("exchange", exchangeName),
			zap.String("symbol", symbol),
			zap.String("unit", string(unit)),
			zap.Int("interval", interval),
		).Error("update future candlestick", zap.Error(err))
	}

	return data
}

func (l *Loader) runDeleteOldRows(ctx context.Context, exchangeName string, oldValueLimit int) {
	_ = scheduler.ExecuteEveryDay(ctx, func() error {
		defer deleteIndicatorHelper()()
		if err := l.priceService.DeleteOldRaws(ctx, exchangeName, time.Now().In(time.UTC).Add(-24*time.Hour)); err != nil {
			l.logger.Error("delete old price", zap.Error(err))
		}
		if err := l.candlestickService.DeleteOldRows(ctx, oldValueLimit); err != nil {
			l.logger.Error("delete old candlestick", zap.Error(err))
		}
		return nil
	})
}
