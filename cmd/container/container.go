package container

import (
	"net/http"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters/exchange"
	infrastracture_http "github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/http"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

var exchangeNames = []string{
	exchange.BinanceExchange,
	exchange.BitgetExchange,
	exchange.BybitExchange,
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
	//Init default logger
	if err := c.di.Provide(func() (*zap.Logger, error) {
		return logger.CreateForNamespace("")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(logger *zap.Logger) metrics.HTTPSender {
		return metrics.NewHTTPSenderWithMetrics(
			infrastracture_http.NewClient(
				http.DefaultClient,
				viper.GetUint64("load.symbols"),
				logger,
			),
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
	_ = c.di.Invoke(func(db *sqlx.DB) {
		_ = db.Close()
	})
	_ = c.di.Invoke(func(conn *amqp.Connection) {
		_ = conn.Close()
	})
	_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.Candlestick]) {
		d.Close()
	})
	_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.Indicator]) {
		d.Close()
	})
	_ = c.di.Invoke(func(d dispatcher.Dispatcher[domain.CreateIndicatorEventBody]) {
		d.Close()
	})
	_ = c.di.Invoke(func(s *http_server.Server) {
		s.Close()
	})

}
