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
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/loader"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/web"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/go-kit/pkg/dispatcher"
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
		serv *service.Service,
		loadedCandleDispatcher dispatcher.Dispatcher[domain.LoadedCandlesticksActionBody],
		candleDispatcher dispatcher.Dispatcher[domain.Candlestick],
		priceDispatcher dispatcher.Dispatcher[domain.LoadedPricesByExchangeActionBody],
		loadedCandleListener dispatcher.Listener[domain.LoadedCandlesticksActionBody],
		candleListener dispatcher.Listener[domain.Candlestick],
		log LoaderLogger,
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
		app.WithLogger(asZapLogger(log))
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
		indicatorService indicator.Indicator,
		analysisService *analysis.Service,
		candleIndicator candle_indicator.CandleIndicator,
		analysisListener dispatcher.Listener[analysis.Analytic],
		candleListener dispatcher.Listener[domain.Candlestick],
		indicatorListener dispatcher.Listener[domain.Indicator],
		h *grpc.ActionHandler,
		log CalculatorLogger,
	) error {
		analyticDispatcher.Subscribe(analysisListener)
		candleDispatcher.Subscribe(candleListener)
		indicatorDispatcher.Subscribe(indicatorListener)
		app = calculator.NewCalculator(
			actionConsumer,
			candlestickConsumer,
			indicatorConsumer,
			analyticConsumer,
			candleIndicatorConsumer,
			analyticDispatcher,
			indicatorDispatcher,
			indicatorService,
			analysisService,
			candleIndicator,
			viper.GetIntSlice("candlestick.depths"),
			viper.GetIntSlice("candlestick.depths"),
		)
		app.WithLogger(asZapLogger(log))
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
		serv *service.Service,
		candleIndicator candle_indicator.CandleIndicator,
		log HTTPServerLogger,
	) error {
		serverHttp = http_server.NewServer()
		serverHttp.AddMiddleware(echoprometheus.NewMiddleware("http_server"))
		serverHttp.WithIndexHtmlResponse(string(web.IndexPage))
		pages, err := web.NewPages()
		if err != nil {
			return err
		}
		serverHttp.RegistrationPage(pages)
		handlerHttp := impl.NewHandler(
			priceService,
			candlestickService,
			indicatorService,
			exchangeService,
			analysisService,
			serv,
			candleIndicator,
		)

		http_server.WithLogger(asZapLogger(log))
		// Лимит вешаем на группу целиком, а не на отдельные адреса: он считается
		// по IP и должен покрывать все методы v1, включая те, что появятся позже.
		apiV1 := serverHttp.ApiGroup("/v1")
		apiV1.Use(impl.RateLimit())
		spec.RegisterHandlers(apiV1, handlerHttp)
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
