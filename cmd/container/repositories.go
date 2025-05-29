package container

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	adapter_repository "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/repository"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func (c *Container) initRepositories() error {
	if err := c.di.Provide(func(conn *sqlx.DB) *postgresql.CandlestickRepository {
		return postgresql.NewCandlestickRepository(conn)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) *postgresql.IndicatorRepository {
		return postgresql.NewIndicatorRepository(conn)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) *postgresql.AnalyticRepository {
		return postgresql.NewAnalyticRepository(conn)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) price.Repository {
		return adapter_repository.NewPriceRepository(postgresql.NewPriceRepository(conn), memory.NewPriceRepository())
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(dbRepo *postgresql.CandlestickRepository) candlestick.Repository {
		return adapter_repository.NewCandlestickRepository(
			dbRepo,
			memory.NewCandlestickRepository(viper.GetInt("candlestick.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(dbRepo *postgresql.IndicatorRepository) indicator.Repository {
		return adapter_repository.NewIndicatorRepository(
			dbRepo,
			memory.NewIndicatorRepository(viper.GetInt("indicator.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(db *postgresql.AnalyticRepository) analysis.Repository {
		return adapter_repository.NewAnalysisRepository(
			db, memory.NewAnalysisRepository(100),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) core_exchange.Repository {
		return adapter_repository.NewExchangeRepository(
			postgresql.NewExchangeRepository(conn), memory.NewExchangeRepository(),
		)
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(conn *sqlx.DB) candle_indicator.Repository {
		return adapter_repository.NewCandleIndicatorRepository(
			postgresql.NewCandleIndicatorRepository(conn),
			memory.NewCandleIndicatorRepository(100),
		)
	}); err != nil {
		return err
	}
	return nil
}
