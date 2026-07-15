package container

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters/repository"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func (c *Container) initRepositories() error {
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) candlestick.Repository {
		decorated := decorateCandlestickRepository(conn, log)
		return repository.NewCandlestickRepository(
			decorated,
			memory.NewCandlestickRepository(viper.GetInt("candlestick.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) indicator.Repository {
		decorated := decorateIndicatorRepository(conn, log)
		return repository.NewIndicatorRepository(
			decorated,
			memory.NewIndicatorRepository(viper.GetInt("indicator.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) analysis.Repository {
		decorated := decorateAnalysisRepository(conn, log)
		return repository.NewAnalysisRepository(
			decorated,
			memory.NewAnalysisRepository(5000),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) price.Repository {
		decorated := decoratePriceRepository(conn, log)
		return repository.NewPriceRepository(decorated, memory.NewPriceRepository())
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) core_exchange.Repository {
		return decorateExchangeRepository(conn, log)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) candle_indicator.Repository {
		decorated := decorateCandleIndicatorRepository(conn, log)
		return repository.NewCandleIndicatorRepository(
			decorated,
			memory.NewCandleIndicatorRepository(5000),
		)
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) postgresql.QueueStore {
		return decorateQueueStore(conn, log)
	}); err != nil {
		return err
	}

	return nil
}
