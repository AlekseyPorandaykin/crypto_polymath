package cmd

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/binance"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bitget"
	v5 "github.com/AlekseyPorandaykin/crypto_loader/pkg/bybit/v5"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/gateio"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kraken"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin"
	kukoin_sender "github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin/sender"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/mexc"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/okx"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/config"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/sqlite"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/loader"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/database"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/server/http"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os/signal"
	"syscall"
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
		//Repositories
		priceRepo := sqlite.NewPriceRepository(conn)
		candlestickRepo := sqlite.NewCandlestickRepository(conn)
		indicatorRepo := sqlite.NewIndicatorRepository(conn)

		//Clients
		binanceClient := system.MustInit[*binance.Manager](binance.NewManager(
			viper.GetString("binance.spot_host"),
			viper.GetString("binance.future_host"),
		))
		bitgetClient := system.MustInit[*bitget.Client](bitget.NewClient(viper.GetString("bitget.host")))
		bybitClient := system.MustInit[*v5.Client](v5.DefaultClient(viper.GetString("bybit.host")))
		gateIoClient := system.MustInit[*gateio.Client](gateio.NewClient(viper.GetString("gateio.host")))
		krakenClient := system.MustInit[*kraken.Client](kraken.NewClient(viper.GetString("kraken.host")))
		kukoinClient := system.MustInit[*kucoin.Client](kucoin.NewClient(viper.GetString("kucoin.host"), kukoin_sender.New()))
		mexcClient := system.MustInit[*mexc.Client](mexc.NewClient(viper.GetString("mexc.host")))
		okxClient := system.MustInit[*okx.Client](okx.NewClient(viper.GetString("okx.host")))

		//Exchanges
		binanceExchange := exchange.NewBinance(binanceClient)
		bybitExchange := exchange.NewByBit(bybitClient)
		bitgetExchange := exchange.NewBitget(bitgetClient)
		gateIoExchange := exchange.NewGateIo(gateIoClient)
		krakenExchange := exchange.NewKraken(krakenClient)
		kukoinExchange := exchange.NewKucoin(kukoinClient)
		mexcExchange := exchange.NewMexc(mexcClient)
		okxExchange := exchange.NewOkx(okxClient)

		//Services
		priceService := price.NewCachingPrice(service.NewCachingPrice(), price.NewService(priceRepo))
		priceService.AddLoader(exchange.BinanceExchange, binanceExchange)
		priceService.AddLoader(exchange.BitgetExchange, bitgetExchange)
		priceService.AddLoader(exchange.BybitExchange, bybitExchange)
		priceService.AddLoader(exchange.GateIoExchange, gateIoExchange)
		priceService.AddLoader(exchange.KrakenExchange, krakenExchange)
		priceService.AddLoader(exchange.KucoinExchange, kukoinExchange)
		priceService.AddLoader(exchange.MexcExchange, mexcExchange)
		priceService.AddLoader(exchange.OkxExchange, okxExchange)

		exchangeNames := []string{
			exchange.BinanceExchange,
			exchange.BitgetExchange,
			exchange.BybitExchange,
			exchange.GateIoExchange,
			exchange.KrakenExchange,
			exchange.KucoinExchange,
			exchange.MexcExchange,
			exchange.OkxExchange,
		}
		symbols := []string{"BTCUSDT", "ETHUSDT"}

		candlestickService := candlestick.NewService(candlestickRepo)
		candlestickService.AddLoader(exchange.BybitExchange, bybitExchange)

		indicatorService := indicator.NewService(indicatorRepo, service.NewCandlestickAdapter(candlestickService))

		loaderApp := loader.NewLoader(priceService, candlestickService, exchangeNames, symbols)
		calculatorApp := calculator.NewCalculator(candlestickService, indicatorService, symbols)

		//Server HTTP
		serverHttp := http.NewServer()
		defer serverHttp.Close()
		serverHttp.AddMiddleware(echoprometheus.NewMiddleware("http_server"))
		handlerHttp := impl.NewHandler(priceService, candlestickService, indicatorService)
		spec.RegisterHandlers(serverHttp.ApiGroup(), handlerHttp)

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

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
