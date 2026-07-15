package container

import (
	"time"

	"github.com/AlekseyPorandaykin/crypto_loader/pkg/binance"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bitget"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/gateio"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kraken"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin"
	kukoin_sender "github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin/sender"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/mexc"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/okx"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/connection"
	"github.com/cenkalti/backoff/v4"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func normalizeDBHost(host string) string {
	if host == "" || host == "0.0.0.0" {
		return "127.0.0.1"
	}
	return host
}

func connectDB(conf connection.DBConfig) (*sqlx.DB, error) {
	var db *sqlx.DB
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.MaxElapsedTime = 30 * time.Second
	err := backoff.Retry(func() error {
		conn, err := connection.CreateDBConnection(conf)
		if err != nil {
			return err
		}
		db = conn
		return nil
	}, bo)
	return db, err
}

func (c *Container) initConnections() error {
	if err := c.di.Provide(func() (*sqlx.DB, error) {
		conf := connection.DBConfig{
			Driver:             viper.GetString("db_connection.driver"),
			Username:           viper.GetString("db_connection.username"),
			Password:           viper.GetString("db_connection.password"),
			Host:               normalizeDBHost(viper.GetString("db_connection.host")),
			Port:               viper.GetString("db_connection.port"),
			Database:           viper.GetString("db_connection.database"),
			PathToDB:           viper.GetString("db_connection.path_to_db"),
			SchemaName:         viper.GetString("db_connection.schema"),
			MaxOpenConnections: viper.GetInt("db_connection.max_open_connections"),
			MaxIdleConnections: viper.GetInt("db_connection.max_idle_connections"),
		}
		zap.L().Info("Create database connection", zap.Any("driver", conf))
		return connectDB(conf)
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
