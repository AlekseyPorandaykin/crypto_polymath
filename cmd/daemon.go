package cmd

import (
	"context"
	"fmt"
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
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	indicator_calculator "github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters"
	adapter_exchange "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	adapter_repository "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/repository"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/config"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/event/listeners"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/sqlite"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/daemon/loader"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/database"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	http_server "github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		defer system.HandlePanic()
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		conf := config.Create()
		conn, errConn := database.CreateConnection(conf.DBConnection)
		if errConn != nil {
			fmt.Println("error create connection", errConn.Error())
			return
		}
		defer func() { _ = conn.Close() }()

		//Constants
		exchangeNames := []string{
			adapter_exchange.BinanceExchange,
			adapter_exchange.BitgetExchange,
			adapter_exchange.BybitExchange,
			adapter_exchange.GateIoExchange,
			adapter_exchange.KrakenExchange,
			adapter_exchange.KucoinExchange,
			adapter_exchange.MexcExchange,
			adapter_exchange.OkxExchange,
		}

		httpClient := metrics.NewHTTPSenderWithMetrics(http.DefaultClient)

		//Repositories
		priceRepo := adapter_repository.NewPriceRepository(sqlite.NewPriceRepository(conn), memory.NewPriceRepository())
		candlestickRepo := adapter_repository.NewCandlestickRepository(
			sqlite.NewCandlestickRepository(conn),
			memory.NewCandlestickRepository(viper.GetInt("candlestick.storage.limit")),
		)
		indicatorDBRepos := sqlite.NewIndicatorRepository(conn)
		indicatorRepo := adapter_repository.NewIndicatorRepository(
			indicatorDBRepos,
			memory.NewIndicatorRepository(viper.GetInt("indicator.storage.limit")),
		)

		//Clients
		binanceClient := system.MustInit[*binance.Manager](binance.NewManager(
			viper.GetString("binance.spot_host"),
			viper.GetString("binance.future_host"),
		))
		bitgetClient := system.MustInit[*bitget.Client](bitget.NewClient(viper.GetString("bitget.host")))
		bybitBasicSender := bybit_sender.NewBasic()
		bybitBasicSender.WithHttpClient(httpClient)

		bybitClient := system.MustInit[*v5.Client](v5.NewClient(viper.GetString("bybit.host"), bybitBasicSender))
		gateIoClient := system.MustInit[*gateio.Client](gateio.NewClient(viper.GetString("gateio.host")))
		krakenClient := system.MustInit[*kraken.Client](kraken.NewClient(viper.GetString("kraken.host")))
		kukoinClient := system.MustInit[*kucoin.Client](kucoin.NewClient(viper.GetString("kucoin.host"), kukoin_sender.New()))
		mexcClient := system.MustInit[*mexc.Client](mexc.NewClient(viper.GetString("mexc.host")))
		okxClient := system.MustInit[*okx.Client](okx.NewClient(viper.GetString("okx.host")))

		//Exchanges
		binanceExchange := adapter_exchange.NewBinance(binanceClient)
		bybitExchange := adapter_exchange.NewByBit(bybitClient)
		bitgetExchange := adapter_exchange.NewBitget(bitgetClient)
		gateIoExchange := adapter_exchange.NewGateIo(gateIoClient)
		krakenExchange := adapter_exchange.NewKraken(krakenClient)
		kukoinExchange := adapter_exchange.NewKucoin(kukoinClient)
		mexcExchange := adapter_exchange.NewMexc(mexcClient)
		okxExchange := adapter_exchange.NewOkx(okxClient)

		//Services
		priceService := price.NewService(priceRepo)
		priceService.AddLoader(adapter_exchange.BinanceExchange, binanceExchange)
		priceService.AddLoader(adapter_exchange.BitgetExchange, bitgetExchange)
		priceService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)
		priceService.AddLoader(adapter_exchange.GateIoExchange, gateIoExchange)
		priceService.AddLoader(adapter_exchange.KrakenExchange, krakenExchange)
		priceService.AddLoader(adapter_exchange.KucoinExchange, kukoinExchange)
		priceService.AddLoader(adapter_exchange.MexcExchange, mexcExchange)
		priceService.AddLoader(adapter_exchange.OkxExchange, okxExchange)

		candlestickService := candlestick.NewService(candlestickRepo)
		candlestickService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)

		indicatorService := indicator.NewService(indicatorRepo, adapters.NewCandlestickAdapter(candlestickService))
		indicatorService.AddCalculator(indicator_calculator.NewMA())
		indicatorService.AddCalculator(indicator_calculator.NewEMA())
		indicatorService.AddCalculator(indicator_calculator.NewTypeCandle())
		indicatorService.AddCalculator(indicator_calculator.NewVolatilityCandlePercent())
		indicatorService.AddCalculator(indicator_calculator.NewTrend())
		indicatorService.AddCalculator(indicator_calculator.NewPriceChanges())
		indicatorService.AddCalculator(indicator_calculator.NewStochasticMainLine())

		exchangeService := core_exchange.New(
			adapter_repository.NewExchangeRepository(sqlite.NewExchangeRepository(conn), memory.NewExchangeRepository()),
		)
		exchangeService.AddLoader(adapter_exchange.BybitExchange, bybitExchange)

		analysisDBRepo := sqlite.NewAnalyticRepository(conn)
		analysisService := analysis.NewService(
			adapter_repository.NewAnalysisRepository(analysisDBRepo, memory.NewAnalysisRepository(100)),
		)

		emaCalc := calculators.NewTrendByEMA(indicatorService, viper.GetIntSlice("candlestick.depths"))
		maCalc := calculators.NewTrendByMA(indicatorService, viper.GetIntSlice("candlestick.depths"))
		rationCandlerToMACalc := calculators.NewRationCandleToMA(candlestickService)
		rationCandlerToEMACalc := calculators.NewRationCandleToEMA(candlestickService)
		rsiCalc := calculators.NewRSI(indicatorService, viper.GetIntSlice("candlestick.depths"))
		macdMainLineCalc := calculators.NewMACDMainLine(indicatorService)
		stochasticSignalLineCalc := calculators.NewStochasticSignalLine(indicatorService)
		analysisService.AddCalculatorByIndicator(emaCalc)
		analysisService.AddCalculatorByIndicator(maCalc)
		analysisService.AddCalculatorByIndicator(rationCandlerToMACalc)
		analysisService.AddCalculatorByIndicator(rationCandlerToEMACalc)
		analysisService.AddCalculatorByIndicator(rsiCalc)
		analysisService.AddCalculatorByIndicator(macdMainLineCalc)
		analysisService.AddCalculatorByIndicator(stochasticSignalLineCalc)

		macdSignalLineCalc := calculators.NewMACDSignalLine(analysisService)
		macdHistogramCalc := calculators.NewMACDHistogram(analysisService)
		analysisService.AddCalculatorByAnalytic(macdSignalLineCalc)
		analysisService.AddCalculatorByAnalytic(macdHistogramCalc)

		//Events
		candleDispatcher := dispatcher.New[domain.Candlestick]()

		defer candleDispatcher.Close()
		indicatorDispatcher := dispatcher.New[domain.Indicator]()
		defer indicatorDispatcher.Close()
		createIndicatorDispatcher := dispatcher.New[domain.CreateIndicatorEventBody]()
		defer createIndicatorDispatcher.Close()
		analyticDispatcher := dispatcher.New[analysis.Analytic]()
		defer analyticDispatcher.Close()

		//Handlers
		indicatorHandler := service.NewIndicatorHandler(
			indicatorService,
			indicatorDispatcher, viper.GetIntSlice("candlestick.depths"),
		)
		analysisHandler := service.NewAnalysisHandler(analysisService, analyticDispatcher)

		candlestickListener := listeners.NewCandlestick(indicatorHandler)
		candleDispatcher.Subscribe(candlestickListener)
		indicatorListener := listeners.NewIndicator(analysisHandler)
		indicatorDispatcher.Subscribe(indicatorListener)

		createIndicatorListener := listeners.NewCreateIndicator(indicatorHandler)
		createIndicatorDispatcher.Subscribe(createIndicatorListener)

		analyticListener := listeners.NewAnalytic(analysisHandler)
		analyticDispatcher.Subscribe(analyticListener)

		//Applications
		loaderApp := loader.NewLoader(
			priceService, candlestickService, exchangeService, indicatorService, candleDispatcher, exchangeNames, viper.GetStringSlice("load.symbols"),
		)
		calculatorApp := calculator.NewCalculator(createIndicatorDispatcher, indicatorService, analysisService, viper.GetStringSlice("load.symbols"))

		//Server HTTP
		serverHttp := http_server.NewServer()
		defer serverHttp.Close()
		serverHttp.AddMiddleware(echoprometheus.NewMiddleware("http_server"))
		handlerHttp := impl.NewHandler(
			priceService,
			candlestickService,
			indicatorService,
			exchangeService,
			analysisService,
			analysisDBRepo,
			indicatorDBRepos,
		)
		spec.RegisterHandlers(serverHttp.ApiGroup(), handlerHttp)

		//Debug config
		if viper.GetBool("app.debug") {
			candleDispatcher.SetPostHandler(func(e dispatcher.Event[domain.Candlestick], duration time.Duration) {
				zap.L().Debug(
					"handle candlestick",
					zap.String("symbol", e.Body.Symbol),
					zap.String("unit", string(e.Body.Unit)),
					zap.Int("interval", e.Body.Interval),
					zap.Time("start_time", e.Body.StartTime),
					zap.Int64("duration_ms", duration.Milliseconds()),
				)
			})
			indicatorDispatcher.SetPostHandler(func(e dispatcher.Event[domain.Indicator], duration time.Duration) {
				zap.L().Debug(
					"handle indicator",
					zap.String("symbol", e.Body.Symbol),
					zap.String("unit", string(e.Body.Unit)),
					zap.Int("interval", e.Body.Interval),
					zap.Time("datetime", e.Body.Datetime),
					zap.String("name", e.Body.Name),
					zap.Int("depth", e.Body.Depth),
					zap.Int64("duration_ms", duration.Milliseconds()),
				)
			})
			analyticDispatcher.SetPostHandler(func(e dispatcher.Event[analysis.Analytic], duration time.Duration) {
				zap.L().Debug("handle analytic",
					zap.String("Symbol", e.Body.Symbol),
					zap.String("Unit", string(e.Body.Unit)),
					zap.Int("Interval", e.Body.Interval),
					zap.String("Name", e.Body.Name),
					zap.Time("datetime", e.Body.Datetime),
					zap.Int("depth", e.Body.Depth),
					zap.Int64("duration_ms", duration.Milliseconds()),
				)
			})
		}

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := loaderApp.Run(ctx); err != nil {
				fmt.Println("run loader app", err.Error())
				return
			}
		})
		system.Go(func() {
			defer cancel()
			if err := calculatorApp.Run(ctx); err != nil {
				fmt.Println("run calculator app", err.Error())
				return
			}
		})
		system.Go(func() {
			defer cancel()
			if err := serverHttp.Run(viper.GetString("http.host"), viper.GetString("http.port")); err != nil {
				fmt.Println("run http server", err.Error())
				return
			}
		})
		system.Go(func() {
			candleDispatcher.Listen()
		})
		system.Go(func() {
			indicatorDispatcher.Listen()
		})
		system.Go(func() {
			createIndicatorDispatcher.Listen()
		})
		system.Go(func() {
			analyticDispatcher.Listen()
		})

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
