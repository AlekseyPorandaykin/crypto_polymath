package container

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/loader"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/web"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/viper"
)

func (c *Container) CreateLoader() (*loader.Loader, error) {
	var app *loader.Loader
	err := c.di.Invoke(func(
		priceService price.Price,
		candlestickService candlestick.Candlestick,
		exchangeService core_exchange.Exchange,
		indicatorService indicator.Indicator,
		serv *application.Service,
		loadedCandleDispatcher dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody],
		candleDispatcher dispatcher.Dispatcher[domain.Candlestick],
		priceDispatcher dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody],
		loadedCandleListener dispatcher.Listener[domain.LoadedCandlesticksActionBody],
		candleListener dispatcher.Listener[domain.Candlestick],
	) error {
		loadedCandleDispatcher.Subscribe(loadedCandleListener)
		candleDispatcher.Subscribe(candleListener)
		app = loader.NewLoader(
			priceService,
			candlestickService,
			exchangeService,
			indicatorService,
			serv,
			loadedCandleDispatcher,
			priceDispatcher,
			candleDispatcher,
			exchangeNames,
			viper.GetStringSlice("load.symbols"),
			viper.GetStringSlice("load.hot_symbols"),
		)
		log, err := logger.CreateForNamespace("loader")
		if err != nil {
			return err
		}
		app.WithLogger(log)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (c *Container) CreateCalculator() (*calculator.Calculator, error) {
	var app *calculator.Calculator
	err := c.di.Invoke(func(
		actionConsumer *postgresql.QueueRepository[queue_contract.Action],
		candlestickConsumer *postgresql.QueueRepository[queue_contract.Candlestick],
		indicatorConsumer *postgresql.QueueRepository[queue_contract.Indicator],
		analyticConsumer *postgresql.QueueRepository[queue_contract.Analytic],
		candleIndicatorConsumer *postgresql.QueueRepository[queue_contract.CandleIndicator],
		analyticDispatcher dispatcher.Dispatcher[analysis.Analytic],
		candleDispatcher dispatcher.Dispatcher[domain.Candlestick],
		indicatorDispatcher dispatcher.Dispatcher[domain.Indicator],
		candleIndicatorDispatcher dispatcher.Dispatcher[candle_indicator.Indicator],
		indicatorService indicator.Indicator,
		analysisService *analysis.Service,
		candleIndicator candle_indicator.CandleIndicator,
		analysisListener dispatcher.Listener[analysis.Analytic],
		candleListener dispatcher.Listener[domain.Candlestick],
		indicatorListener dispatcher.Listener[domain.Indicator],
		candleIndicatorListener dispatcher.Listener[candle_indicator.Indicator],
		h *grpc.ActionHandler,
	) error {
		analyticDispatcher.Subscribe(analysisListener)
		candleDispatcher.Subscribe(candleListener)
		indicatorDispatcher.Subscribe(indicatorListener)
		candleIndicatorDispatcher.Subscribe(candleIndicatorListener)
		app = calculator.NewCalculator(
			actionConsumer,
			candlestickConsumer,
			indicatorConsumer,
			analyticConsumer,
			candleIndicatorConsumer,
			analyticDispatcher,
			indicatorDispatcher,
			candleIndicatorDispatcher,
			indicatorService,
			analysisService,
			candleIndicator,
			viper.GetIntSlice("candlestick.depths"),
			viper.GetIntSlice("candlestick.depths"),
		)
		log, err := logger.CreateForNamespace("calculator")
		if err != nil {
			return err
		}
		app.WithLogger(log)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (c *Container) CreateGRPCServer() (*grpc.Server, error) {
	var server *grpc.Server
	err := c.di.Invoke(func(h *grpc.ActionHandler) error {
		grpcServer, err := grpc.NewServer(viper.GetUint("grpc.port"), h)
		if err != nil {
			return err
		}
		server = grpcServer
		return nil

	})
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *Container) CreateApiServer() (*http_server.Server, error) {
	var serverHttp *http_server.Server
	err := c.di.Invoke(func(
		priceService price.Price,
		candlestickService candlestick.Candlestick,
		indicatorService indicator.Indicator,
		exchangeService core_exchange.Exchange,
		analysisService *analysis.Service,
		analysisDBRepo *postgresql.AnalyticRepository,
		indicatorDBRepos *postgresql.IndicatorRepository,
		serv *application.Service,
		candleIndicator candle_indicator.CandleIndicator,
	) error {
		serverHttp = http_server.NewServer()
		serverHttp.AddMiddleware(echoprometheus.NewMiddleware("http_server"))
		serverHttp.WithIndexHtmlResponse(string(web.IndexPage))
		handlerHttp := impl.NewHandler(
			priceService,
			candlestickService,
			indicatorService,
			exchangeService,
			analysisService,
			analysisDBRepo,
			indicatorDBRepos,
			serv,
			candleIndicator,
		)

		log, err := logger.CreateForNamespace("http_server")
		if err != nil {
			return err
		}
		http_server.WithLogger(log)
		spec.RegisterHandlers(serverHttp.ApiGroup("/v1"), handlerHttp)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return serverHttp, nil
}

func (c *Container) CreateCandleDispatcher() (dispatcher.Dispatcher[domain.Candlestick], error) {
	var d dispatcher.Dispatcher[domain.Candlestick]
	err := c.di.Invoke(func(candleDispatcher dispatcher.Dispatcher[domain.Candlestick]) {
		d = candleDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateIndicatorDispatcher() (dispatcher.Dispatcher[domain.Indicator], error) {
	var d dispatcher.Dispatcher[domain.Indicator]
	err := c.di.Invoke(func(indicatorDispatcher dispatcher.Dispatcher[domain.Indicator]) {
		d = indicatorDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateCreatorIndicatorDispatcher() (dispatcher.Dispatcher[domain.CreateIndicatorEventBody], error) {
	var d dispatcher.Dispatcher[domain.CreateIndicatorEventBody]
	err := c.di.Invoke(func(createIndicatorDispatcher dispatcher.Dispatcher[domain.CreateIndicatorEventBody]) {
		d = createIndicatorDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateLoadedCandlesticksForSymbolBodyDispatcher() (dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody], error) {
	var d dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody]
	err := c.di.Invoke(func(sourceDispatcher dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody]) {
		d = sourceDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateAnalyticDispatcher() (dispatcher.Dispatcher[analysis.Analytic], error) {
	var d dispatcher.Dispatcher[analysis.Analytic]
	err := c.di.Invoke(func(analyticDispatcher dispatcher.Dispatcher[analysis.Analytic]) {
		d = analyticDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}
func (c *Container) CreateLoadedPricesDispatcher() (dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody], error) {
	var d dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody]
	err := c.di.Invoke(func(priceDispatcher dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody]) {
		d = priceDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}
func (c *Container) CreateCandleIndicatorsDispatcher() (dispatcher.Dispatcher[candle_indicator.Indicator], error) {
	var d dispatcher.Dispatcher[candle_indicator.Indicator]
	err := c.di.Invoke(func(ciDispatcher dispatcher.Dispatcher[candle_indicator.Indicator]) {
		d = ciDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}
