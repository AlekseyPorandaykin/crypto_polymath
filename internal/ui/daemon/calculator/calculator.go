package calculator

import (
	"context"
	"errors"
	"fmt"
	"time"

	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-template/pkg/scheduler"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//Обработка
// - analysis : CalculateByIndicator,
// - candle_indicator : CalculateFromCandlesticks, CalculateAllIndicators
// - indicator : CalcIndicatorsByCandlestick

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

	analyticDispatcher  dispatcher.Dispatcher[analysis.Analytic]
	indicatorDispatcher dispatcher.Dispatcher[domain.Indicator]

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
	indicatorService indicator.Indicator,
	analysisService *analysis.Service,
	candleIndicator candle_indicator.CandleIndicator,
	indicatorDepths []int,
	candlestickDepths []int,
) *Calculator {
	return &Calculator{
		actionConsumer:          actionConsumer,
		candlestickConsumer:     candlestickConsumer,
		indicatorConsumer:       indicatorConsumer,
		analyticConsumer:        analyticConsumer,
		candleIndicatorConsumer: candleIndicatorConsumer,
		analyticDispatcher:      analyticDispatcher,
		indicatorDispatcher:     indicatorDispatcher,
		indicatorService:        indicatorService,
		analysisService:         analysisService,
		candleIndicator:         candleIndicator,
		indicatorDepths:         indicatorDepths,
		candlestickDepths:       candlestickDepths,
		logger:                  zap.L().Named("calculator"),
	}
}

func (app *Calculator) WithLogger(logger *zap.Logger) {
	app.logger = logger
}

func (app *Calculator) Run(ctx context.Context) error {
	errCh := make(chan error, 10)

	system.Go(func() {
		app.analyticDispatcher.Listen()
	})
	system.Go(func() {
		app.indicatorDispatcher.Listen()
	})
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
			defer deleteIndicatorHelper()()
			if err := app.indicatorService.DeleteOldRows(
				util.AddRequestID(ctx),
				viper.GetInt("indicator.storage.limit"),
			); err != nil {
				app.logger.Error(
					"delete old indicators",
					zap.Error(err),
					zap.String("request_id", util.RequestIDFromContext(ctx)),
				)
			}
			return nil
		})
	})
	system.Go(func() {
		_ = scheduler.ExecuteEveryHour(ctx, 1, 1, func() error {
			defer deleteAnalysisHelper()()
			if err := app.analysisService.DeleteOldRows(
				util.AddRequestID(ctx),
				viper.GetInt("analysis.storage.limit"),
			); err != nil {
				app.logger.Error("delete old analysis", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
			return nil
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
			messages, err := app.actionConsumer.Receive(ctx)
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
				continue
			}
			if len(messages) == 0 {
				continue
			}
			messagesGrouped := make(map[string][]*queue_contract.Action, len(messages))
			for _, message := range messages {
				key := fmt.Sprintf("%s-%s-%s-%d", message.Exchange, message.Symbol, message.Unit, message.Interval)
				if _, ok := messagesGrouped[key]; !ok {
					messagesGrouped[key] = make([]*queue_contract.Action, 0, 100)
				}
				messagesGrouped[key] = append(messagesGrouped[key], message)
			}
			g := errgroup.Group{}
			g.SetLimit(50)
			for _, messagesGroups := range messagesGrouped {
				mg := messagesGroups
				g.Go(func() error {
					if errH := app.handleAction(util.AddRequestID(ctx), mg); errH != nil {
						return errH
					}
					return nil
				})
			}
			if errW := g.Wait(); errW != nil {
				app.logger.Error("failed to handle action", zap.Error(errW), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
		}
	}
}
func (app *Calculator) listenCandlestick(ctx context.Context) error {
	defer app.candlestickConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			messages, err := app.candlestickConsumer.Receive(ctx)
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
				continue
			}
			if len(messages) == 0 {
				continue
			}
			messagesGrouped := make(map[string][]*queue_contract.Candlestick, len(messages))
			for _, message := range messages {
				key := fmt.Sprintf("%s-%s-%s-%d", message.Exchange, message.Symbol, message.Unit, message.Interval)
				if _, ok := messagesGrouped[key]; !ok {
					messagesGrouped[key] = make([]*queue_contract.Candlestick, 0, 100)
				}
				messagesGrouped[key] = append(messagesGrouped[key], message)
			}
			g := errgroup.Group{}
			g.SetLimit(50)
			for _, messagesGroups := range messagesGrouped {
				mg := messagesGroups
				g.Go(func() error {
					if errH := app.handleCandlestick(util.AddRequestID(ctx), mg); errH != nil {
						return errH
					}
					return nil
				})
			}
			if errW := g.Wait(); errW != nil {
				app.logger.Error("failed to handle action", zap.Error(errW), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
		}
	}
}
func (app *Calculator) listenIndicators(ctx context.Context) error {
	defer app.indicatorConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			messages, err := app.indicatorConsumer.Receive(ctx)
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
				continue
			}
			if len(messages) == 0 {
				continue
			}
			messagesGrouped := make(map[string][]*queue_contract.Indicator, len(messages))
			for _, message := range messages {
				key := fmt.Sprintf("%s-%s-%s-%d", message.Exchange, message.Symbol, message.Unit, message.Interval)
				if _, ok := messagesGrouped[key]; !ok {
					messagesGrouped[key] = make([]*queue_contract.Indicator, 0, 100)
				}
				messagesGrouped[key] = append(messagesGrouped[key], message)
			}
			g := errgroup.Group{}
			g.SetLimit(50)
			for _, messagesGroups := range messagesGrouped {
				mg := messagesGroups
				g.Go(func() error {
					if errH := app.handleIndicators(util.AddRequestID(ctx), mg); errH != nil {
						return errH
					}
					return nil
				})
			}
			if errW := g.Wait(); errW != nil {
				app.logger.Error("failed to handle indicator", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
		}
	}
}

type analyticGroupKey struct {
	exchange string
	symbol   string
	unit     string
	interval int
}

func (app *Calculator) listenAnalytic(ctx context.Context) error {
	defer app.analyticConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			messages, err := app.analyticConsumer.Receive(ctx)
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
				continue
			}
			if len(messages) == 0 {
				continue
			}
			messagesGrouped := make(map[analyticGroupKey][]*queue_contract.Analytic, len(messages))
			for _, message := range messages {
				key := analyticGroupKey{
					exchange: message.Exchange,
					symbol:   message.Symbol,
					unit:     message.Unit,
					interval: message.Interval,
				}
				messagesGrouped[key] = append(messagesGrouped[key], message)
			}
			g := errgroup.Group{}
			g.SetLimit(50)
			for _, messagesGroups := range messagesGrouped {
				mg := messagesGroups
				g.Go(func() error {
					if errH := app.handleAnalytic(util.AddRequestID(ctx), mg); errH != nil {
						return errH
					}
					return nil
				})
			}
			if errW := g.Wait(); errW != nil {
				app.logger.Error("failed to handle analytic", zap.Error(errW), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
		}
	}
}

// Нет потребителя на это событие
func (app *Calculator) listenCandleIndicatorConsumer(ctx context.Context) error {
	defer app.candleIndicatorConsumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			messages, err := app.candleIndicatorConsumer.Receive(ctx)
			if errors.Is(err, queue.ErrClosed) {
				return nil
			}
			if err != nil {
				app.logger.Error("failed to receive message", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
				continue
			}
			if len(messages) == 0 {
				continue
			}
			if err := app.handleCandleIndicatorConsumer(util.AddRequestID(ctx), messages); err != nil {
				app.logger.Error("failed to handle action", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
			}
		}
	}
}

func (app *Calculator) handleAction(ctx context.Context, messages []*queue_contract.Action) error {
	//TODO: надо создавать задачи в очередь
	slice.SortBy(messages, func(a, b *queue_contract.Action) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	})
	start := time.Now()
	for _, m := range messages {
		if m.Name != domain.LoadedCandlesticksForSymbolAction {
			app.logger.Debug("received unsupported action", zap.String("action", m.Name), zap.String("request_id", util.RequestIDFromContext(ctx)))
			continue
		}
		action := domain.LoadedCandlesticksActionBody{
			Exchange:  m.Exchange,
			Symbol:    m.Symbol,
			Unit:      domain.Unit(m.Unit),
			Interval:  m.Interval,
			CreatedAt: m.CreatedAt,
			Duration:  m.Duration,
		}
		if action.Unit == domain.MinuteUnit || action.Interval != 1 {
			// Пропускаем минутные свечи и не единичные, т.к. они не нужны для расчета индикаторов.
			return nil
		}
		_, err := app.candleIndicator.CalculateAllIndicators(
			ctx,
			action.Exchange,
			action.Symbol,
			action.Unit,
			action.Interval,
		)
		if err != nil {
			app.logger.Error("calculate all indicators", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
		}
		g := errgroup.Group{}
		g.SetLimit(20)
		for _, depth := range app.indicatorDepths {
			d := depth
			g.Go(func() error {
				indicatorData, errC := app.indicatorService.CalcIndicators(
					ctx,
					action.Exchange,
					action.Symbol,
					action.Unit,
					action.Interval,
					d)
				if errC != nil {
					return errC
				}
				for i := range indicatorData {
					app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
						Name: domain.CreatedIndicatorEvent,
						Body: indicatorData[i],
					})
				}
				return nil
			})
		}
		if errW := g.Wait(); errW != nil {
			app.logger.Error("failed to handle action", zap.Error(errW), zap.String("request_id", util.RequestIDFromContext(ctx)))
		}
	}
	if len(messages) > 0 {
		first := messages[0]
		app.logger.Info(
			"Handle LoadedCandlesticksForSymbolAction event",
			zap.String("duration", time.Since(start).String()),
			zap.String("Exchange", first.Exchange),
			zap.String("Symbol", first.Symbol),
			zap.String("Unit", first.Unit),
			zap.Int("Interval", first.Interval),
			zap.String("Datetime", first.CreatedAt.String()),
			zap.Int("Count", len(messages)),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}
	return nil
}
func (app *Calculator) handleCandlestick(ctx context.Context, messages []*queue_contract.Candlestick) error {
	slice.SortBy(messages, func(a, b *queue_contract.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	start := time.Now()
	for _, m := range messages {
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
		_, err := app.candleIndicator.CalculateFromCandlesticks(ctx, []domain.Candlestick{candle})
		if err != nil {
			app.logger.Error("calculate candle indicator", zap.Error(err), zap.Any("candle", candle), zap.String("request_id", util.RequestIDFromContext(ctx)))
		}
		app.calculateIndicatorsByCandle(ctx, candle)

	}
	if len(messages) > 0 {
		first := messages[0]
		app.logger.Info(
			"Handle candlestick event",
			zap.String("duration", time.Since(start).String()),
			zap.String("Exchange", first.Exchange),
			zap.String("Symbol", first.Symbol),
			zap.String("Unit", first.Unit),
			zap.String("Unit", first.Unit),
			zap.Int("Interval", first.Interval),
			zap.String("StartTime", first.StartTime.String()),
			zap.Int("Count", len(messages)),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}

	return nil
}
func (app *Calculator) handleIndicators(ctx context.Context, messages []*queue_contract.Indicator) error {
	slice.SortBy(messages, func(a, b *queue_contract.Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})
	start := time.Now()
	for _, m := range messages {
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
			app.logger.Error("calculate indicators", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
			continue
		}
		for _, analyticItem := range analytics {
			app.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
				Name: domain.CreatedAnalyticEvent,
				Body: analyticItem,
			})
		}
	}
	if len(messages) > 0 {
		first := messages[0]
		app.logger.Info(
			"Handle indicator event",
			zap.String("duration", time.Since(start).String()),
			zap.String("Exchange", first.Exchange),
			zap.String("Symbol", first.Symbol),
			zap.String("Unit", first.Unit),
			zap.Int("Interval", first.Interval),
			zap.String("Datetime", first.Datetime.String()),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
			zap.Int("Count", len(messages)),
		)
	}

	return nil
}
func (app *Calculator) handleAnalytic(ctx context.Context, messages []*queue_contract.Analytic) error {
	start := time.Now()
	slice.SortBy(messages, func(a, b *queue_contract.Analytic) bool {
		return a.Datetime.Before(b.Datetime)
	})
	sourceAnalytics := make([]analysis.Analytic, len(messages))
	for i, m := range messages {
		sourceAnalytics[i] = analysis.Analytic{
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
	}
	analytics, err := app.analysisService.CalculateByAnalytics(ctx, sourceAnalytics)
	if err != nil {
		app.logger.Error("calculate analytics", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
		return nil
	}
	for _, analyticItem := range analytics {
		app.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
			Name: domain.CreatedAnalyticEvent,
			Body: analyticItem,
		})
	}
	if len(messages) > 0 {
		first := messages[0]
		app.logger.Debug(
			"Handle analytic batch",
			zap.String("duration", time.Since(start).String()),
			zap.Int("messages", len(messages)),
			zap.Int("produced", len(analytics)),
			zap.String("Exchange", first.Exchange),
			zap.String("Symbol", first.Symbol),
			zap.String("Unit", first.Unit),
			zap.Int("Interval", first.Interval),
			zap.Int("count", len(messages)),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}
	return nil
}
func (app *Calculator) handleCandleIndicatorConsumer(ctx context.Context, messages []*queue_contract.CandleIndicator) error {
	return nil
}

func (app *Calculator) calculateIndicatorsByCandle(ctx context.Context, c domain.Candlestick) {
	start := time.Now()
	g := errgroup.Group{}
	g.SetLimit(20)
	for _, depth := range app.candlestickDepths {
		d := depth
		g.Go(func() error {
			data, err := app.indicatorService.CalcIndicatorsByCandlestick(ctx, c, d)
			if err != nil {
				return err
			}
			for _, item := range data {
				app.indicatorDispatcher.Dispatch(dispatcher.Event[domain.Indicator]{
					Name: domain.CreatedIndicatorEvent,
					Body: item,
				})
			}
			return nil
		})

	}
	if err := g.Wait(); err != nil {
		app.logger.Error("calculate indicators", zap.Error(err), zap.String("request_id", util.RequestIDFromContext(ctx)))
	}
	app.logger.Info(
		"indicator calculated",
		zap.String("duration", time.Since(start).String()), zap.String("request_id", util.RequestIDFromContext(ctx)))
}
