package loader_test

import (
	"context"
	"fmt"
	"testing"

	crypto_exchage_config "github.com/AlekseyPorandaykin/crypto-exchanges/config"
	"github.com/AlekseyPorandaykin/crypto-exchanges/factory"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	indicator_calculator "github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/config"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters"
	repository2 "github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters/repository"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/sqlite"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-kit/pkg/connection"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/viper"
)

func init() {
	viper.Set("app.codename", "crypto_polymath")
	viper.Set("http.host", "")
	viper.Set("http.port", "8085")
	viper.Set("price.duration.loader", "30s")

	viper.Set("bybit.host", "https://api.bybit.com/")
	viper.Set("binance.spot_host", "https://api.binance.com")
	viper.Set("binance.future_host", "https://fapi.binance.com")
	viper.Set("bitget.host", "https://api.bitget.com/")
	viper.Set("kucoin.host", "https://api.kucoin.com/")
	viper.Set("okx.host", "https://www.okx.com/")
	viper.Set("gateio.host", "https://api.gateio.ws/")
	viper.Set("kraken.host", "https://api.kraken.com/")
	viper.Set("mexc.host", "https://api.mexc.com/")

	viper.Set("db_connection.driver", "sqlite")
	viper.Set("db_connection.path_to_db", "/Users/alexey.porandaikin/Projects/go/projects/crypto_polymath/storage/crypto_polymath.db")

	viper.Set("candlestick.minutes", []int{1, 15, 30})
	viper.Set("candlestick.hours", []int{1, 2, 4, 6, 12})
	viper.Set("candlestick.depths", []int{1, 10, 20, 50})
	viper.Set("candlestick.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
}
func TestLoader_Run(t *testing.T) {
	ctx := context.TODO()
	conf := config.Create()
	conn, errConn := connection.CreateDBConnection(conf.DBConnection)
	if errConn != nil {
		fmt.Println("error create connection", errConn.Error())
		return
	}
	defer func() { _ = conn.Close() }()
	//symbols := []string{"BTCUSDT", "ETHUSDT"}
	candlestickRepo := repository2.NewCandlestickRepository(
		sqlite.NewCandlestickRepository(conn),
		memory.NewCandlestickRepository(viper.GetInt("candlestick.storage.limit")),
	)
	indicatorRepo := repository2.NewIndicatorRepository(
		sqlite.NewIndicatorRepository(conn),
		memory.NewIndicatorRepository(viper.GetInt("indicator.storage.limit")),
	)
	bybitExchange := system.MustInit(factory.NewExchangeClient(crypto_exchage_config.BybitV5Config{
		BaseUrl:            "https://api.bybit.com/",
		AllowLogger:        true,
		AllowRequestLogger: true,
		AllowWaitAdder:     true,
	}))
	bybit := adapters.NewExchange(bybitExchange)
	candlestickService := candlestick.NewService(candlestickRepo)
	candlestickService.AddLoader(bybit.Name(), bybit)
	indicatorService := indicator.NewService(indicatorRepo, adapters.NewCandlestickAdapter(candlestickService))
	indicatorService.AddCalculator(indicator_calculator.NewMA())
	indicatorService.AddCalculator(indicator_calculator.NewEMA())
	indicatorService.AddCalculator(indicator_calculator.NewTypeCandle())
	indicatorService.AddCalculator(indicator_calculator.NewVolatilityCandlePercent())
	indicatorService.AddCalculator(indicator_calculator.NewTrend())
	indicatorService.AddCalculator(indicator_calculator.NewPriceChanges())
	candlestickService.SequenceCandlesticks(ctx, bybit.Name(), "BTCUSDT", domain.MonthUnit, 1, 100)
	data, err := candlestickService.LoadCandlesticks(ctx, bybit.Name(), "BTCUSDT", domain.MonthUnit, 1)
	if err != nil {
		fmt.Println(errConn.Error())
		return
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	indicators := make([]domain.Indicator, 0, 100)
	for _, item := range data {
		primaryIndicators, err := indicatorService.CalcIndicatorsByCandlestick(ctx, item, 10)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		indicators = append(indicators, primaryIndicators...)
	}
	slice.SortBy[domain.Indicator](indicators, func(a, b domain.Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})

	fmt.Println(data, err)
}
