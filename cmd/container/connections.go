package container

import (
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/binance"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bitget"
	v5 "github.com/AlekseyPorandaykin/crypto_loader/pkg/bybit/v5"
	bybit_sender "github.com/AlekseyPorandaykin/crypto_loader/pkg/bybit/v5/sender"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/gateio"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kraken"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin"
	kukoin_sender "github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin/sender"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/mexc"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/okx"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/connection"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func (c *Container) initConnections() error {
	if err := c.di.Provide(func() (*sqlx.DB, error) {
		conf := connection.DBConfig{
			Driver:             viper.GetString("db_connection.driver"),
			Username:           viper.GetString("db_connection.username"),
			Password:           viper.GetString("db_connection.password"),
			Host:               viper.GetString("db_connection.host"),
			Port:               viper.GetString("db_connection.port"),
			Database:           viper.GetString("db_connection.database"),
			PathToDB:           viper.GetString("db_connection.path_to_db"),
			SchemaName:         viper.GetString("db_connection.schema"),
			MaxOpenConnections: 5,
			MaxIdleConnections: 5,
		}
		zap.L().Info("Create database connection", zap.Any("driver", conf))
		return connection.CreateDBConnection(conf)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*amqp.Connection, error) {
		return queue.CreateRabbitConnection(
			viper.GetString("rabbit_mq.username"),
			viper.GetString("rabbit_mq.password"),
			viper.GetString("rabbit_mq.addr"),
		)
	}); err != nil {
		return err
	}
	return nil
}

func (c *Container) initClients() error {
	if err := c.di.Provide(func() (*binance.Manager, error) {
		return binance.NewManager(
			viper.GetString("binance.spot_host"),
			viper.GetString("binance.future_host"),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*bitget.Client, error) {
		return bitget.NewClient(viper.GetString("bitget.host"))
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(httpClient metrics.HTTPSender) (*v5.Client, error) {
		bybitBasicSender := bybit_sender.NewBasic()
		bybitBasicSender.WithHttpClient(httpClient)
		log, err := logger.CreateForNamespace("bybit_sender")
		if err != nil {
			return nil, err
		}
		bybitSender := bybit_sender.NewWaitAdder(bybit_sender.NewRequestLogger(bybitBasicSender, log))
		client, err := v5.NewClient(viper.GetString("bybit.host"), bybitSender)
		if err != nil {
			return nil, err
		}
		client.WithLogger(log)
		return client, nil
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*gateio.Client, error) {
		return gateio.NewClient(viper.GetString("gateio.host"))
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*kraken.Client, error) {
		return kraken.NewClient(viper.GetString("kraken.host"))
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*kucoin.Client, error) {
		return kucoin.NewClient(viper.GetString("kucoin.host"), kukoin_sender.New())
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*mexc.Client, error) {
		return mexc.NewClient(viper.GetString("mexc.host"))
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (*okx.Client, error) {
		return okx.NewClient(viper.GetString("okx.host"))
	}); err != nil {
		return err
	}
	return nil
}
