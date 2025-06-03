package calculator

import (
	"context"
	"errors"
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-template/pkg/scheduler"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

//TODO: Запускаем и считываем курсором новые свечи, проверяем какая свеча и запускаем пайплайн(event-source) расчета индикаторов.
// Все значения запуска будут видны здесь.

// Calculator -	компонент, который отвечает за расчет индикаторов и аналитики.
// Запускается по в виде демона и выполняет расчеты по расписанию.
type Calculator struct {
	actionConsumer          queue.Receiver[queue_contract.Action]
	candlestickConsumer     queue.Receiver[queue_contract.Candlestick]
	indicatorConsumer       queue.Receiver[queue_contract.Indicator]
	analyticConsumer        queue.Receiver[queue_contract.Analytic]
	candleIndicatorConsumer queue.Receiver[queue_contract.CandleIndicator]

	analyticDispatcher        dispatcher.Dispatcher[analysis.Analytic]
	indicatorDispatcher       dispatcher.Dispatcher[domain.Indicator]
	candleIndicatorDispatcher dispatcher.Dispatcher[candle_indicator.Indicator]

	clientSender     *grpc.ActionHandler
	indicatorService indicator.Indicator
	analysisService  *analysis.Service
	candleIndicator  candle_indicator.CandleIndicator

	indicatorDepths   []int
	candlestickDepths []int
	logger            *zap.Logger
}

func NewCalculator(
	actionConsumer queue.Receiver[queue_contract.Action],
	candlestickConsumer queue.Receiver[queue_contract.Candlestick],
	indicatorConsumer queue.Receiver[queue_contract.Indicator],
	analyticConsumer queue.Receiver[queue_contract.Analytic],
	candleIndicatorConsumer queue.Receiver[queue_contract.CandleIndicator],
	analyticDispatcher dispatcher.Dispatcher[analysis.Analytic],
	indicatorDispatcher dispatcher.Dispatcher[domain.Indicator],
	candleIndicatorDispatcher dispatcher.Dispatcher[candle_indicator.Indicator],
	clientSender *grpc.ActionHandler,
	indicatorService indicator.Indicator,
	analysisService *analysis.Service,
	candleIndicator candle_indicator.CandleIndicator,
	indicatorDepths []int,
	candlestickDepths []int,
) *Calculator {
	return &Calculator{
		actionConsumer:            actionConsumer,
		candlestickConsumer:       candlestickConsumer,
		indicatorConsumer:         indicatorConsumer,
		analyticConsumer:          analyticConsumer,
		candleIndicatorConsumer:   candleIndicatorConsumer,
		analyticDispatcher:        analyticDispatcher,
		indicatorDispatcher:       indicatorDispatcher,
		candleIndicatorDispatcher: candleIndicatorDispatcher,
		clientSender:              clientSender,
		indicatorService:          indicatorService,
		analysisService:           analysisService,
		candleIndicator:           candleIndicator,
		indicatorDepths:           indicatorDepths,
		candlestickDepths:         candlestickDepths,
		logger:                    zap.L(),
	}
}

func (app *Calculator) WithLogger(logger *zap.Logger) {
	app.logger = logger
}

func (app *Calculator) Run(ctx context.Context) error {
	errCh := make(chan error)
	system.Go(func() {
		app.actionConsumer.Listen()
	})
	system.Go(func() {
		app.candlestickConsumer.Listen()
	})
	system.Go(func() {
		app.indicatorConsumer.Listen()
	})
	system.Go(func() {
		app.analyticConsumer.Listen()
	})
	system.Go(func() {
		app.candleIndicatorConsumer.Listen()
	})
	system.Go(func() {
		if err := app.listenAction(ctx); err != nil {
			errCh <- err
		}
	})
	system.Go(func() {
		if err := app.listenCandlestick(ctx); err != nil {
			errCh <- err
		}
	})
	system.Go(func() {
		if err := app.listenIndicators(ctx); err != nil {
			errCh <- err
		}
	})
	system.Go(func() {
		if err := app.listenAnalytic(ctx); err != nil {
			errCh <- err
		}
	})
	system.Go(func() {
		if err := app.listenCandleIndicatorConsumer(ctx); err != nil {
			errCh <- err
		}
	})

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

func (app *Calculator) listenAction(ctx context.Context) error {
	defer app.actionConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := app.actionConsumer.Receive()
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err))
				continue
			}
			if message == nil {
				continue
			}
			if err := app.handleAction(ctx, *message); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err))
			}
		}
	}
}
func (app *Calculator) listenCandlestick(ctx context.Context) error {
	defer app.actionConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := app.candlestickConsumer.Receive()
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err))
				continue
			}
			if message == nil {
				continue
			}
			if err := app.handleCandlestick(ctx, *message); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err))
			}
		}
	}
}
func (app *Calculator) listenIndicators(ctx context.Context) error {
	defer app.actionConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := app.indicatorConsumer.Receive()
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err))
				continue
			}
			if message == nil {
				continue
			}
			if err := app.handleIndicators(ctx, *message); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err))
			}
		}
	}
}
func (app *Calculator) listenAnalytic(ctx context.Context) error {
	defer app.actionConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := app.analyticConsumer.Receive()
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err))
				continue
			}
			if message == nil {
				continue
			}
			if err := app.handleAnalytic(ctx, *message); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err))
			}
		}
	}
}
func (app *Calculator) listenCandleIndicatorConsumer(ctx context.Context) error {
	defer app.actionConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := app.candleIndicatorConsumer.Receive()
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err))
				continue
			}
			if message == nil {
				continue
			}
			if err := app.handleCandleIndicatorConsumer(ctx, *message); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err))
			}
		}
	}
}

func (app *Calculator) handleAction(ctx context.Context, m queue_contract.Action) error {
	switch m.Name {
	case domain.LoadedCandlesticksForSymbolAction:
		action := domain.LoadedCandlesticksActionBody{
			Exchange:  m.Exchange,
			Symbol:    m.Symbol,
			Unit:      domain.Unit(m.Unit),
			Interval:  m.Interval,
			CreatedAt: m.CreatedAt,
			Duration:  m.Duration,
		}
		app.clientSender.Accept(m.Name, action)
		if action.Unit == domain.MinuteUnit || action.Interval != 1 {
			// Пропускаем минутные свечи и не единичные, т.к. они не нужны для расчета индикаторов.
			return nil
		}
		candleIndicators, err := app.candleIndicator.CalculateAllIndicators(
			ctx,
			action.Exchange,
			action.Symbol,
			action.Unit,
			action.Interval,
		)
		if err != nil {
			app.logger.Error("calculate all indicators", zap.Error(err))
		}
		for i := range candleIndicators {
			app.candleIndicatorDispatcher.Dispatch(dispatcher.Event[candle_indicator.Indicator]{
				Name: domain.CreatedIndicatorEvent,
				Body: candleIndicators[i],
			})
		}
		for _, depth := range app.indicatorDepths {
			indicatorData, err := app.indicatorService.CalcIndicators(
				ctx,
				action.Exchange,
				action.Symbol,
				action.Unit,
				action.Interval,
				depth)
			if err != nil {
				app.logger.Error("calculate indicators", zap.Error(err), zap.Any("action", action))
				continue
			}
			for i := range indicatorData {
				app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
					Name: domain.CreatedIndicatorEvent,
					Body: indicatorData[i],
				})
			}
		}
	default:
		app.logger.Debug("received unsupported action", zap.String("action", m.Name))
	}
	return nil
}
func (app *Calculator) handleCandlestick(ctx context.Context, m queue_contract.Candlestick) error {
	candle := domain.Candlestick{
		Exchange:   m.Exchange,
		Symbol:     m.Symbol,
		Unit:       domain.Unit(m.Unit),
		Interval:   m.Interval,
		StartTime:  m.StartTime,
		OpenPrice:  m.OpenPrice,
		HighPrice:  m.HighPrice,
		LowPrice:   m.LowPrice,
		ClosePrice: m.ClosePrice,
		Volume:     m.Volume,
	}
	if candle.Unit == domain.MinuteUnit || candle.Interval != 1 {
		//Пропускаем минутные свечи и не единичные, т.к. они не нужны для расчета индикаторов.
		return nil
	}
	candleIndicatorsData, err := app.candleIndicator.CalculateFromCandlesticks(ctx, []domain.Candlestick{candle})
	if err != nil {
		app.logger.Error("calculate candle indicator", zap.Error(err), zap.Any("candle", candle))
	}
	for i := range candleIndicatorsData {
		app.candleIndicatorDispatcher.Dispatch(dispatcher.Event[candle_indicator.Indicator]{
			Name: domain.CreatedIndicatorEvent,
			Body: candleIndicatorsData[i],
		})
	}
	//TODO рассчитывем на основе события , а не свечи
	//app.calculateIndicatorsByCandle(ctx, candle)

	return nil
}
func (app *Calculator) handleIndicators(ctx context.Context, m queue_contract.Indicator) error {
	indicatorData := domain.Indicator{
		Exchange: m.Exchange,
		Symbol:   m.Symbol,
		Unit:     domain.Unit(m.Unit),
		Interval: m.Interval,
		Datetime: m.Datetime,
		Name:     m.Name,
		Depth:    m.Depth,
		Value:    m.Value,
	}
	analytics, err := app.analysisService.CalculateByIndicator(ctx, indicatorData)
	if err != nil {
		return err
	}
	for _, analyticItem := range analytics {
		app.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
			Name: domain.CreatedAnalyticEvent,
			Body: analyticItem,
		})
	}
	return nil
}
func (app *Calculator) handleAnalytic(ctx context.Context, m queue_contract.Analytic) error {
	data := analysis.Analytic{
		ID:             m.ID,
		Exchange:       m.Exchange,
		Symbol:         m.Symbol,
		Unit:           domain.Unit(m.Unit),
		Interval:       m.Interval,
		Name:           m.Name,
		Datetime:       m.Datetime,
		Depth:          m.Depth,
		ByIndicator:    m.ByIndicator,
		IndicatorDepth: m.IndicatorDepth,
		Value:          m.Value,
	}
	analytics, err := app.analysisService.CalculateByAnalytic(ctx, data)
	if err != nil {
		return err
	}
	for _, analyticItem := range analytics {
		app.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
			Name: domain.CreatedAnalyticEvent,
			Body: analyticItem,
		})
	}
	return nil
}
func (app *Calculator) handleCandleIndicatorConsumer(ctx context.Context, m queue_contract.CandleIndicator) error {
	_ = m
	return nil
}

func (app *Calculator) calculateIndicatorsByCandle(ctx context.Context, c domain.Candlestick) {
	start := time.Now()
	indicators := make([]domain.Indicator, 0, len(app.candlestickDepths))
	for _, depth := range app.candlestickDepths {
		data, err := app.indicatorService.CalcIndicators(ctx, c.Exchange, c.Symbol, c.Unit, c.Interval, depth)
		if err != nil {
			app.logger.Error("calculate indicator", zap.Error(err))
			continue
		}
		for _, item := range data {
			app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
				Name: domain.CreatedIndicatorEvent,
				Body: item,
			})
		}
		indicators = append(indicators, data...)
	}
	app.logger.Debug(
		"indicator calculated",
		zap.Int("count", len(indicators)),
		zap.String("duration", time.Since(start).String()))
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
