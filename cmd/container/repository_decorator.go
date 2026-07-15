package container

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	observabilitylogging "github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/observability/logging"
	observabilityprometheus "github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/observability/prometheus"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/jmoiron/sqlx"
)

const repositoryDB = "crypto_polymath"

func decorateCandleIndicatorRepository(conn *sqlx.DB, logger RepositoryLogger) candle_indicator.Repository {
	var repo candle_indicator.Repository = postgresql.NewCandleIndicatorRepository(conn)
	repo = observabilityprometheus.NewCandleIndicatorRepository(repo, repositoryDB)
	return observabilitylogging.NewCandleIndicatorRepository(repo, asZapLogger(logger), repositoryDB)
}

func decorateAnalysisRepository(conn *sqlx.DB, logger RepositoryLogger) analysis.Repository {
	var repo analysis.Repository = postgresql.NewAnalyticRepository(conn)
	repo = observabilityprometheus.NewAnalysisRepository(repo, repositoryDB)
	return observabilitylogging.NewAnalysisRepository(repo, asZapLogger(logger), repositoryDB)
}

func decorateCandlestickRepository(conn *sqlx.DB, logger RepositoryLogger) candlestick.Repository {
	var repo candlestick.Repository = postgresql.NewCandlestickRepository(conn)
	repo = observabilityprometheus.NewCandlestickRepository(repo, repositoryDB)
	return observabilitylogging.NewCandlestickRepository(repo, asZapLogger(logger), repositoryDB)
}

func decorateIndicatorRepository(conn *sqlx.DB, logger RepositoryLogger) indicator.Repository {
	var repo indicator.Repository = postgresql.NewIndicatorRepository(conn)
	repo = observabilityprometheus.NewIndicatorRepository(repo, repositoryDB)
	return observabilitylogging.NewIndicatorRepository(repo, asZapLogger(logger), repositoryDB)
}

func decorateExchangeRepository(conn *sqlx.DB, logger RepositoryLogger) core_exchange.Repository {
	var repo core_exchange.Repository = postgresql.NewExchangeRepository(conn)
	repo = observabilityprometheus.NewExchangeRepository(repo, repositoryDB)
	return observabilitylogging.NewExchangeRepository(repo, asZapLogger(logger), repositoryDB)
}

func decoratePriceRepository(conn *sqlx.DB, logger RepositoryLogger) price.Repository {
	var repo price.Repository = postgresql.NewPriceRepository(conn)
	repo = observabilityprometheus.NewPriceRepository(repo, repositoryDB)
	return observabilitylogging.NewPriceRepository(repo, asZapLogger(logger), repositoryDB)
}

func decorateQueueStore(conn *sqlx.DB, logger RepositoryLogger) postgresql.QueueStore {
	var store postgresql.QueueStore = postgresql.NewQueueStore(conn)
	store = observabilityprometheus.NewQueueStore(store, repositoryDB)
	return observabilitylogging.NewQueueStore(store, asZapLogger(logger), repositoryDB)
}
