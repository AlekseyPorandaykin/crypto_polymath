package container

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/event/listeners"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"time"
)

func (c *Container) initEventImplementations() error {
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQProducer[queue_contract.Candlestick], error) {
		producer, err := queue.NewRabbitMQProducer[queue_contract.Candlestick](
			conn,
			viper.GetString("events.queue_candlestick"),
		)
		if err != nil {
			return nil, err
		}
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQProducer[queue_contract.Action], error) {
		producer, err := queue.NewRabbitMQProducer[queue_contract.Action](conn, viper.GetString("events.queue_action"))
		if err != nil {
			return nil, err
		}
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQProducer[queue_contract.Indicator], error) {
		producer, err := queue.NewRabbitMQProducer[queue_contract.Indicator](conn, viper.GetString("events.queue_indicator"))
		if err != nil {
			return nil, err
		}
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQProducer[queue_contract.Analytic], error) {
		producer, err := queue.NewRabbitMQProducer[queue_contract.Analytic](conn, viper.GetString("events.queue_analytic"))
		if err != nil {
			return nil, err
		}
		return producer, nil
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQProducer[queue_contract.CandleIndicator], error) {
		producer, err := queue.NewRabbitMQProducer[queue_contract.CandleIndicator](conn, viper.GetString("events.queue_candle_indicator"))
		if err != nil {
			return nil, err
		}
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQConsumer[queue_contract.Action], error) {
		consumer, err := queue.NewRabbitMQConsumer[queue_contract.Action](
			conn,
			viper.GetString("events.queue_action"),
			viper.GetString("app.codename"),
		)
		if err != nil {
			return nil, err
		}
		return consumer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQConsumer[queue_contract.Candlestick], error) {
		consumer, err := queue.NewRabbitMQConsumer[queue_contract.Candlestick](
			conn,
			viper.GetString("events.queue_candlestick"),
			viper.GetString("app.codename"),
		)
		if err != nil {
			return nil, err
		}
		return consumer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQConsumer[queue_contract.Indicator], error) {
		consumer, err := queue.NewRabbitMQConsumer[queue_contract.Indicator](
			conn,
			viper.GetString("events.queue_indicator"),
			viper.GetString("app.codename"),
		)
		if err != nil {
			return nil, err
		}
		return consumer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQConsumer[queue_contract.Analytic], error) {
		consumer, err := queue.NewRabbitMQConsumer[queue_contract.Analytic](
			conn,
			viper.GetString("events.queue_analytic"),
			viper.GetString("app.codename"),
		)
		if err != nil {
			return nil, err
		}
		return consumer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *amqp.Connection) (*queue.RabbitMQConsumer[queue_contract.CandleIndicator], error) {
		consumer, err := queue.NewRabbitMQConsumer[queue_contract.CandleIndicator](
			conn,
			viper.GetString("events.queue_candle_indicator"),
			viper.GetString("app.codename"),
		)
		if err != nil {
			return nil, err
		}
		return consumer, nil
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(store postgresql.QueueStore) (*postgresql.QueueRepository[queue_contract.CandleIndicator], error) {
		producer := postgresql.NewQueueRepository[queue_contract.CandleIndicator](
			store,
			viper.GetString("events.queue_candle_indicator"),
			2*time.Hour,
			viper.GetInt("events.queue_batch_size"),
		)
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(store postgresql.QueueStore) (*postgresql.QueueRepository[queue_contract.Analytic], error) {
		producer := postgresql.NewQueueRepository[queue_contract.Analytic](
			store,
			viper.GetString("events.queue_analytic"),
			2*time.Hour,
			viper.GetInt("events.queue_batch_size"),
		)
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(store postgresql.QueueStore) (*postgresql.QueueRepository[queue_contract.Indicator], error) {
		producer := postgresql.NewQueueRepository[queue_contract.Indicator](
			store,
			viper.GetString("events.queue_indicator"),
			2*time.Hour,
			viper.GetInt("events.queue_batch_size"),
		)
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(store postgresql.QueueStore) (*postgresql.QueueRepository[queue_contract.Action], error) {
		producer := postgresql.NewQueueRepository[queue_contract.Action](
			store,
			viper.GetString("events.queue_action"),
			2*time.Hour,
			viper.GetInt("events.queue_batch_size"),
		)
		return producer, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(store postgresql.QueueStore) (*postgresql.QueueRepository[queue_contract.Candlestick], error) {
		producer := postgresql.NewQueueRepository[queue_contract.Candlestick](
			store,
			viper.GetString("events.queue_candlestick"),
			2*time.Hour,
			viper.GetInt("events.queue_batch_size"),
		)
		return producer, nil
	}); err != nil {
		return err
	}

	return nil
}

func (c *Container) initListeners() error {
	//Listeners
	if err := c.di.Provide(func(
		p *postgresql.QueueRepository[queue_contract.Action],
		clientSender *grpc.ActionHandler,
	) dispatcher.Listener[domain.LoadedCandlesticksActionBody] {
		return listeners.NewLoadedCandlesticks(p, clientSender)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		p *postgresql.QueueRepository[queue_contract.Candlestick],
	) dispatcher.Listener[domain.Candlestick] {
		return listeners.NewCandlestick(p)
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(
		p *postgresql.QueueRepository[queue_contract.Indicator],
	) dispatcher.Listener[domain.Indicator] {
		return listeners.NewIndicator(p)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		p *postgresql.QueueRepository[queue_contract.Analytic],
	) dispatcher.Listener[analysis.Analytic] {
		return listeners.NewAnalytic(p)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		p *postgresql.QueueRepository[queue_contract.CandleIndicator],
	) dispatcher.Listener[candle_indicator.Indicator] {
		return listeners.NewCandleIndicator(p)
	}); err != nil {
		return err
	}
	return nil
}

func (c *Container) initEvents() error {
	if err := c.di.Provide(func() dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody] {
		candleDispatcher := dispatcher.NewMemoryDispatcher[domain.LoadedCandlesticksActionBody]()
		candleDispatcher.SetPreDispatcher(func(e dispatcher.Event[domain.LoadedCandlesticksActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		candleDispatcher.SetPostDispatcher(func(e dispatcher.Event[domain.LoadedCandlesticksActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		candleDispatcher.SetPreHandler(func(e dispatcher.Event[domain.LoadedCandlesticksActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		candleDispatcher.SetPostHandler(func(e dispatcher.Event[domain.LoadedCandlesticksActionBody], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return candleDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[domain.Candlestick] {
		candleDispatcher := dispatcher.NewMemoryDispatcher[domain.Candlestick]()
		candleDispatcher.SetPreDispatcher(func(e dispatcher.Event[domain.Candlestick]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		candleDispatcher.SetPostDispatcher(func(e dispatcher.Event[domain.Candlestick]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		candleDispatcher.SetPreHandler(func(e dispatcher.Event[domain.Candlestick]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		candleDispatcher.SetPostHandler(func(e dispatcher.Event[domain.Candlestick], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return candleDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[domain.Indicator] {
		indicatorDispatcher := dispatcher.NewMemoryDispatcher[domain.Indicator]()
		indicatorDispatcher.SetPreDispatcher(func(e dispatcher.Event[domain.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		indicatorDispatcher.SetPostDispatcher(func(e dispatcher.Event[domain.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		indicatorDispatcher.SetPreHandler(func(e dispatcher.Event[domain.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		indicatorDispatcher.SetPostHandler(func(e dispatcher.Event[domain.Indicator], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues("indicator").Add(float64(duration.Milliseconds()))
		})
		return indicatorDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[domain.CreateIndicatorEventBody] {
		createIndicatorDispatcher := dispatcher.NewEmptyDispatcher[domain.CreateIndicatorEventBody]()
		createIndicatorDispatcher.SetPreDispatcher(func(e dispatcher.Event[domain.CreateIndicatorEventBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		createIndicatorDispatcher.SetPostDispatcher(func(e dispatcher.Event[domain.CreateIndicatorEventBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		createIndicatorDispatcher.SetPreHandler(func(e dispatcher.Event[domain.CreateIndicatorEventBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		createIndicatorDispatcher.SetPostHandler(func(e dispatcher.Event[domain.CreateIndicatorEventBody], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return createIndicatorDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[analysis.Analytic] {
		analyticDispatcher := dispatcher.NewMemoryDispatcher[analysis.Analytic]()
		analyticDispatcher.SetPreDispatcher(func(e dispatcher.Event[analysis.Analytic]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		analyticDispatcher.SetPostDispatcher(func(e dispatcher.Event[analysis.Analytic]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		analyticDispatcher.SetPreHandler(func(e dispatcher.Event[analysis.Analytic]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		analyticDispatcher.SetPostHandler(func(e dispatcher.Event[analysis.Analytic], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return analyticDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody] {
		analyticDispatcher := dispatcher.NewEmptyDispatcher[domain.LoadedPricesByExchangeActionBody]()
		analyticDispatcher.SetPreDispatcher(func(e dispatcher.Event[domain.LoadedPricesByExchangeActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		analyticDispatcher.SetPostDispatcher(func(e dispatcher.Event[domain.LoadedPricesByExchangeActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		analyticDispatcher.SetPreHandler(func(e dispatcher.Event[domain.LoadedPricesByExchangeActionBody]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		analyticDispatcher.SetPostHandler(func(e dispatcher.Event[domain.LoadedPricesByExchangeActionBody], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return analyticDispatcher
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() dispatcher.Dispatcher[candle_indicator.Indicator] {
		analyticDispatcher := dispatcher.NewMemoryDispatcher[candle_indicator.Indicator]()
		analyticDispatcher.SetPreDispatcher(func(e dispatcher.Event[candle_indicator.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "add").Inc()
		})
		analyticDispatcher.SetPostDispatcher(func(e dispatcher.Event[candle_indicator.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "added").Inc()
		})
		analyticDispatcher.SetPreHandler(func(e dispatcher.Event[candle_indicator.Indicator]) {
			metrics.EventCount.WithLabelValues(e.Name, "handle").Inc()
		})
		analyticDispatcher.SetPostHandler(func(e dispatcher.Event[candle_indicator.Indicator], duration time.Duration) {
			metrics.EventCount.WithLabelValues(e.Name, "handled").Inc()
			metrics.EventDurationQuery.WithLabelValues(e.Name).Add(float64(duration.Milliseconds()))
		})
		return analyticDispatcher
	}); err != nil {
		return err
	}
	return nil

}
