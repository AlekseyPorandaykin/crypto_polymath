package container

import (
	"net/http"

	"github.com/AlekseyPorandaykin/crypto-exchanges/client"
	"github.com/AlekseyPorandaykin/crypto-exchanges/config"
	"github.com/AlekseyPorandaykin/crypto-exchanges/factory"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/binance"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bitget"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/gateio"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kraken"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin"
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
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters/exchange"
	grpc2 "github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/service"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func (c *Container) initExchanges() error {
	if err := c.di.Provide(func(binanceClient *binance.Manager) *exchange.Binance {
		return exchange.NewBinance(binanceClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(rt http.RoundTripper, log HTTPClientLogger) (client.ExchangeClient, error) {
		return factory.NewExchangeClient(config.BybitV5Config{
			BaseUrl:            viper.GetString("bybit.host"),
			AllowLogger:        true,
			AllowRequestLogger: true,
			AllowWaitAdder:     true,
		}, factory.WithLogger(asZapLogger(log)), factory.WithHttpRoundTripper(rt))
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(bybitClient client.ExchangeClient) *adapters.Exchange {
		return adapters.NewExchange(bybitClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(bitgetClient *bitget.Client) *exchange.Bitget {
		return exchange.NewBitget(bitgetClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(gateIoClient *gateio.Client) *exchange.GateIo {
		return exchange.NewGateIo(gateIoClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(krakenClient *kraken.Client) *exchange.Kraken {
		return exchange.NewKraken(krakenClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(kukoinClient *kucoin.Client) *exchange.Kucoin {
		return exchange.NewKucoin(kukoinClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(mexcClient *mexc.Client) *exchange.Mexc {
		return exchange.NewMexc(mexcClient)
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func(okxClient *okx.Client) *exchange.Okx {
		return exchange.NewOkx(okxClient)
	}); err != nil {
		return err
	}
	return nil
}

func (c *Container) initServices() error {
	if err := c.di.Provide(func(priceRepo price.Repository) price.Price {
		priceService := price.NewService(priceRepo)
		err := c.di.Invoke(func(
			binanceExchange *exchange.Binance,
			bitgetExchange *exchange.Bitget,
			bybitExchange *adapters.Exchange,
			gateIoExchange *exchange.GateIo,
			krakenExchange *exchange.Kraken,
			kukoinExchange *exchange.Kucoin,
			mexcExchange *exchange.Mexc,
			okxExchange *exchange.Okx,
		) {
			priceService.AddLoader(exchange.BinanceExchange, binanceExchange)
			priceService.AddLoader(exchange.BitgetExchange, bitgetExchange)
			priceService.AddLoader(bybitExchange.Name(), bybitExchange)
			priceService.AddLoader(exchange.GateIoExchange, gateIoExchange)
			priceService.AddLoader(exchange.KrakenExchange, krakenExchange)
			priceService.AddLoader(exchange.KucoinExchange, kukoinExchange)
			priceService.AddLoader(exchange.MexcExchange, mexcExchange)
			priceService.AddLoader(exchange.OkxExchange, okxExchange)
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
		bybitExchange *adapters.Exchange,
	) candlestick.Candlestick {
		candlestickService := candlestick.NewService(candlestickRepo)
		candlestickService.AddLoader(bybitExchange.Name(), bybitExchange)
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
		bybitExchange *adapters.Exchange,
		repo core_exchange.Repository,
	) core_exchange.Exchange {
		exchangeService := core_exchange.New(repo)
		exchangeService.AddLoader(bybitExchange.Name(), bybitExchange)
		return exchangeService
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(
		indicatorService indicator.Indicator, candlestickService candlestick.Candlestick, repo analysis.Repository,
	) *analysis.Service {
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

	if err := c.di.Provide(func(conn *sqlx.DB, log RepositoryLogger) service.DictionaryRepositories {
		return service.DictionaryRepositories{
			Analysis:   decorateAnalysisRepository(conn, log),
			Indicators: decorateIndicatorRepository(conn, log),
			Symbols:    decorateCandlestickRepository(conn, log),
		}
	}); err != nil {
		return err
	}

	if err := c.di.Provide(func(repos service.DictionaryRepositories) *service.Service {
		return service.NewService(repos)
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

	if err := c.di.Provide(func() *grpc2.ActionHandler {
		return grpc2.NewActionHandler()
	}); err != nil {
		return err
	}
	return nil
}
