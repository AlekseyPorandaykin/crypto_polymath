package cmd

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
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis/calculators"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	candle_indicator_calc "github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator/calculators"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	indicator_calculator "github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters"
	adapter_exchange "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	adapter_repository "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/repository"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/config"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/event/listeners"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/sqlite"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/loader"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/database"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/logger"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var exchangeNames = []string{
	adapter_exchange.BinanceExchange,
	adapter_exchange.BitgetExchange,
	adapter_exchange.BybitExchange,
	adapter_exchange.GateIoExchange,
	adapter_exchange.KrakenExchange,
	adapter_exchange.KucoinExchange,
	adapter_exchange.MexcExchange,
	adapter_exchange.OkxExchange,
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
	if err := c.di.Provide(func() config.AppConf {
		return config.Create()
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conf config.AppConf) (*sqlx.DB, error) {
		return database.CreateConnection(conf.DBConnection)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() metrics.HTTPSender {
		return metrics.NewHTTPSenderWithMetrics(http.DefaultClient)
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
	if err := c.initEvents(); err != nil {
		return err
	}
	if err := c.initServices(); err != nil {
		return err
	}
	return nil
}

func (c *Container) CreateLoader() (*loader.Loader, error) {
	var app *loader.Loader
	err := c.di.Invoke(func(
		priceService price.Price,
		candlestickService candlestick.Candlestick,
		exchangeService core_exchange.Exchange,
		indicatorService indicator.Indicator,
		candleDispatcher *dispatcher.Dispatcher[domain.Candlestick],
		serv *application.Service,
		candleIndicator candle_indicator.CandleIndicator,
	) error {
		app = loader.NewLoader(
			priceService,
			candlestickService,
			exchangeService,
			indicatorService,
			candleDispatcher,
			serv,
			candleIndicator,
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
		createIndicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody],
		indicatorService indicator.Indicator,
		analysisService *analysis.Service,
	) {
		app = calculator.NewCalculator(
			createIndicatorDispatcher, indicatorService, analysisService, viper.GetStringSlice("load.symbols"),
		)
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (c *Container) CreateApiServer() (*http_server.Server, error) {
	var serverHttp *http_server.Server
	err := c.di.Invoke(func(
		priceService price.Price,
		candlestickService candlestick.Candlestick,
		indicatorService indicator.Indicator,
		exchangeService core_exchange.Exchange,
		analysisService *analysis.Service,
		analysisDBRepo *sqlite.AnalyticRepository,
		indicatorDBRepos *sqlite.IndicatorRepository,
		serv *application.Service,
		candleIndicator candle_indicator.CandleIndicator,
	) {
		serverHttp = http_server.NewServer()
		serverHttp.AddMiddleware(echoprometheus.NewMiddleware("http_server"))
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
		spec.RegisterHandlers(serverHttp.ApiGroup(), handlerHttp)
	})
	if err != nil {
		return nil, err
	}
	return serverHttp, nil
}

func (c *Container) CreateCandleDispatcher() (*dispatcher.Dispatcher[domain.Candlestick], error) {
	var d *dispatcher.Dispatcher[domain.Candlestick]
	err := c.di.Invoke(func(candleDispatcher *dispatcher.Dispatcher[domain.Candlestick]) {
		d = candleDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateIndicatorDispatcher() (*dispatcher.Dispatcher[domain.Indicator], error) {
	var d *dispatcher.Dispatcher[domain.Indicator]
	err := c.di.Invoke(func(indicatorDispatcher *dispatcher.Dispatcher[domain.Indicator]) {
		d = indicatorDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateCreaterIndicatorDispatcher() (*dispatcher.Dispatcher[domain.CreateIndicatorEventBody], error) {
	var d *dispatcher.Dispatcher[domain.CreateIndicatorEventBody]
	err := c.di.Invoke(func(createIndicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody]) {
		d = createIndicatorDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) CreateAnalyticDispatcher() (*dispatcher.Dispatcher[analysis.Analytic], error) {
	var d *dispatcher.Dispatcher[analysis.Analytic]
	err := c.di.Invoke(func(analyticDispatcher *dispatcher.Dispatcher[analysis.Analytic]) {
		d = analyticDispatcher
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Container) initRepositories() error {
	if err := c.di.Provide(func(conn *sqlx.DB) price.Repository {
		return adapter_repository.NewPriceRepository(sqlite.NewPriceRepository(conn), memory.NewPriceRepository())
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) candlestick.Repository {
		return adapter_repository.NewCandlestickRepository(
			sqlite.NewCandlestickRepository(conn),
			memory.NewCandlestickRepository(viper.GetInt("candlestick.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) *sqlite.IndicatorRepository {
		return sqlite.NewIndicatorRepository(conn)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(db *sqlite.IndicatorRepository) indicator.Repository {
		return adapter_repository.NewIndicatorRepository(
			db,
			memory.NewIndicatorRepository(viper.GetInt("indicator.storage.limit")),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) *sqlite.AnalyticRepository {
		return sqlite.NewAnalyticRepository(conn)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(db *sqlite.AnalyticRepository) analysis.Repository {
		return adapter_repository.NewAnalysisRepository(
			db, memory.NewAnalysisRepository(100),
		)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(conn *sqlx.DB) core_exchange.Repository {
		return adapter_repository.NewExchangeRepository(
			sqlite.NewExchangeRepository(conn), memory.NewExchangeRepository(),
		)
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(conn *sqlx.DB) candle_indicator.Repository {
		return adapter_repository.NewCandleIndicatorRepository(
			sqlite.NewCandleIndicatorRepository(conn),
			memory.NewCandleIndicatorRepository(100),
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
		bybitBasicSender.WithLogger(log)
		return v5.NewClient(viper.GetString("bybit.host"), bybitBasicSender)
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

func (c *Container) initExchanges() error {
	if err := c.di.Provide(func(binanceClient *binance.Manager) *adapter_exchange.Binance {
		return adapter_exchange.NewBinance(binanceClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(bybitClient *v5.Client) *adapter_exchange.Bybit {
		return adapter_exchange.NewByBit(bybitClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(bitgetClient *bitget.Client) *adapter_exchange.Bitget {
		return adapter_exchange.NewBitget(bitgetClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(gateIoClient *gateio.Client) *adapter_exchange.GateIo {
		return adapter_exchange.NewGateIo(gateIoClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(krakenClient *kraken.Client) *adapter_exchange.Kraken {
		return adapter_exchange.NewKraken(krakenClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(kukoinClient *kucoin.Client) *adapter_exchange.Kucoin {
		return adapter_exchange.NewKucoin(kukoinClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(mexcClient *mexc.Client) *adapter_exchange.Mexc {
		return adapter_exchange.NewMexc(mexcClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(okxClient *okx.Client) *adapter_exchange.Okx {
		return adapter_exchange.NewOkx(okxClient)
	}); err != nil {
		return err
	}
	return nil
}

func (c *Container) initServices() error {
	if err := c.di.Provide(func(priceRepo price.Repository) price.Price {
		priceService := price.NewService(priceRepo)
		err := c.di.Invoke(func(
			binanceExchange *adapter_exchange.Binance,
			bitgetExchange *adapter_exchange.Bitget,
			bybitExchange *adapter_exchange.Bybit,
			gateIoExchange *adapter_exchange.GateIo,
			krakenExchange *adapter_exchange.Kraken,
			kukoinExchange *adapter_exchange.Kucoin,
			mexcExchange *adapter_exchange.Mexc,
			okxExchange *adapter_exchange.Okx,
		) {
			priceService.AddLoader(adapter_exchange.BinanceExchange, binanceExchange)
			priceService.AddLoader(adapter_exchange.BitgetExchange, bitgetExchange)
			priceService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)
			priceService.AddLoader(adapter_exchange.GateIoExchange, gateIoExchange)
			priceService.AddLoader(adapter_exchange.KrakenExchange, krakenExchange)
			priceService.AddLoader(adapter_exchange.KucoinExchange, kukoinExchange)
			priceService.AddLoader(adapter_exchange.MexcExchange, mexcExchange)
			priceService.AddLoader(adapter_exchange.OkxExchange, okxExchange)
		})
		if err != nil {
			return nil
		}
		return priceService
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		candlestickRepo candlestick.Repository,
		bybitExchange *adapter_exchange.Bybit,
	) candlestick.Candlestick {
		candlestickService := candlestick.NewService(candlestickRepo)
		candlestickService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)
		return candlestickService
	}); err != nil {
		return err
	}
	if err := c.di.Provide(
		func(indicatorRepo indicator.Repository, candlestickService candlestick.Candlestick,
		) indicator.Indicator {
			indicatorService := indicator.NewService(indicatorRepo, adapters.NewCandlestickAdapter(candlestickService))
			indicatorService.AddCalculator(indicator_calculator.NewMA())
			indicatorService.AddCalculator(indicator_calculator.NewEMA())
			indicatorService.AddCalculator(indicator_calculator.NewTypeCandle())
			indicatorService.AddCalculator(indicator_calculator.NewVolatilityCandlePercent())
			indicatorService.AddCalculator(indicator_calculator.NewTrend())
			indicatorService.AddCalculator(indicator_calculator.NewPriceChanges())
			indicatorService.AddCalculator(indicator_calculator.NewStochasticMainLine())
			return indicatorService
		}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		conn *sqlx.DB,
		bybitExchange *adapter_exchange.Bybit,
		repo core_exchange.Repository,
	) core_exchange.Exchange {
		exchangeService := core_exchange.New(repo)
		exchangeService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)
		return exchangeService
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(indicatorService indicator.Indicator, candlestickService candlestick.Candlestick, repo analysis.Repository) *analysis.Service {
		analysisService := analysis.NewService(repo, indicatorService, viper.GetIntSlice("candlestick.depths"))
		analysisService.AddCalculatorByIndicator(calculators.NewTrendByEMA(indicatorService))
		analysisService.AddCalculatorByIndicator(calculators.NewTrendByMA(indicatorService))
		analysisService.AddCalculatorByIndicator(calculators.NewRationCandleToMA(candlestickService))
		analysisService.AddCalculatorByIndicator(calculators.NewRationCandleToEMA(candlestickService))
		analysisService.AddCalculatorByIndicator(calculators.NewRSI(indicatorService))
		analysisService.AddCalculatorByIndicator(calculators.NewMACDMainLine(indicatorService))
		analysisService.AddCalculatorByIndicator(calculators.NewStochasticSignalLine(indicatorService))
		analysisService.AddCalculatorByAnalytic(calculators.NewMACDSignalLine(analysisService))
		analysisService.AddCalculatorByAnalytic(calculators.NewMACDHistogram(analysisService))

		return analysisService
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		indicatorService indicator.Indicator,
		indicatorDispatcher *dispatcher.Dispatcher[domain.Indicator],
		candleDispatcher *dispatcher.Dispatcher[domain.Candlestick],
		createIndicatorDispatcher *dispatcher.Dispatcher[domain.CreateIndicatorEventBody],
		candleIndicator candle_indicator.CandleIndicator,
	) *application.IndicatorHandler {
		indicatorHandler := application.NewIndicatorHandler(
			indicatorService,
			indicatorDispatcher, viper.GetIntSlice("candlestick.depths"),
		)
		indicatorHandler.SetLogger(zap.L())
		candleDispatcher.Subscribe(listeners.NewCandlestick(indicatorHandler, candleIndicator))
		createIndicatorDispatcher.Subscribe(listeners.NewCreateIndicator(indicatorHandler))
		return indicatorHandler
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		analysisService *analysis.Service,
		analyticDispatcher *dispatcher.Dispatcher[analysis.Analytic],
		indicatorDispatcher *dispatcher.Dispatcher[domain.Indicator],
	) *application.AnalysisHandler {
		analysisHandler := application.NewAnalysisHandler(analysisService, analyticDispatcher)
		indicatorDispatcher.Subscribe(listeners.NewIndicator(analysisHandler))
		analyticDispatcher.Subscribe(listeners.NewAnalytic(analysisHandler))
		return analysisHandler
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(
		analysisDBRepo *sqlite.AnalyticRepository, indicatorDBRepos *sqlite.IndicatorRepository,
	) *application.Service {
		return application.NewService(analysisDBRepo, indicatorDBRepos)
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(
		repo candle_indicator.Repository,
		candlestickService candlestick.Candlestick,
	) candle_indicator.CandleIndicator {
		s := candle_indicator.New(repo, candlestickService)
		s.AddCalculator(candle_indicator_calc.NewHeikenAshi(repo))
		return s
	}); err != nil {
		return err
	}

	return nil
}

func (c *Container) initEvents() error {
	if err := c.di.Provide(func() *dispatcher.Dispatcher[domain.Candlestick] {
		candleDispatcher := dispatcher.New[domain.Candlestick]()
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
	if err := c.di.Provide(func() *dispatcher.Dispatcher[domain.Indicator] {
		indicatorDispatcher := dispatcher.New[domain.Indicator]()
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
	if err := c.di.Provide(func() *dispatcher.Dispatcher[domain.CreateIndicatorEventBody] {
		createIndicatorDispatcher := dispatcher.New[domain.CreateIndicatorEventBody]()
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
	if err := c.di.Provide(func() *dispatcher.Dispatcher[analysis.Analytic] {
		analyticDispatcher := dispatcher.New[analysis.Analytic]()
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
	return nil

}

func (c *Container) Close() {
	_ = c.di.Invoke(func(db *sqlx.DB) {
		_ = db.Close()
	})
	_ = c.di.Invoke(func(d *dispatcher.Dispatcher[domain.Candlestick]) {
		d.Close()
	})
	_ = c.di.Invoke(func(d *dispatcher.Dispatcher[domain.Indicator]) {
		d.Close()
	})
	_ = c.di.Invoke(func(d *dispatcher.Dispatcher[domain.CreateIndicatorEventBody]) {
		d.Close()
	})
	_ = c.di.Invoke(func(s *http_server.Server) {
		s.Close()
	})

}
