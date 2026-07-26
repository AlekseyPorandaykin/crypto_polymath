package container

import (
	"net/http"

	v5 "github.com/AlekseyPorandaykin/crypto-exchanges/exchange/bybit/v5"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters/exchange"
	infrastracture_http "github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/http"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/go-kit/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

var exchangeNames = []string{
	exchange.BinanceExchange,
	exchange.BitgetExchange,
	v5.ExchangeName,
	exchange.GateIoExchange,
	exchange.KrakenExchange,
	exchange.KucoinExchange,
	exchange.MexcExchange,
	exchange.OkxExchange,
}

type Container struct {
	di *dig.Container
}

func NewContainer() *Container {
	return &Container{
		di: dig.New(),
	}
}
func (c *Container) Init() error {
	if err := c.initConnections(); err != nil {
		return err
	}
	if err := c.initLoggers(); err != nil {
		return err
	}
	if err := c.di.Provide(func(log AppLogger) metrics.HTTPSender {
		return metrics.NewHTTPSenderWithMetrics(
			infrastracture_http.NewClient(
				http.DefaultClient,
				10,
				asZapLogger(log),
			),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(_ AppLogger) http.RoundTripper {
		return metrics.NewRoundTripperWithMetrics(
			http.DefaultTransport, nil,
		)
	}); err != nil {
		return err
	}
	if err := c.initRepositories(); err != nil {
		return err
	}
	if err := c.initClients(); err != nil {
		return err
	}
	if err := c.initExchanges(); err != nil {
		return err
	}
	if err := c.initEventImplementations(); err != nil {
		return err
	}
	if err := c.initEvents(); err != nil {
		return err
	}
	if err := c.initListeners(); err != nil {
		return err
	}
	if err := c.initServices(); err != nil {
		return err
	}
	return nil
}

func (c *Container) Close() {
	g := errgroup.Group{}
	g.Go(func() error {
		//logger.SyncLoggers()
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.Candlestick]) {
			d.Close()
		})
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.Indicator]) {
			d.Close()
		})
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.CreateIndicatorEventBody]) {
			d.Close()
		})
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(s *http_server.Server) {
			s.Close()
		})
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(db *sqlx.DB) {
			_ = db.Close()
		})
		return nil
	})
	g.Go(func() error {
		_ = c.di.Invoke(func(conn *amqp.Connection) {
			_ = conn.Close()
		})
		return nil
	})
	_ = g.Wait()
}
