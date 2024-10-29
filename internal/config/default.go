package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.Set("app.codename", "crypto_polymath")
	viper.Set("app.debug", false)
	viper.Set("http.host", "")
	viper.Set("http.port", "80")
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
	viper.Set("db_connection.path_to_db", "./storage/crypto_polymath.db")

	viper.Set("candlestick.minutes", []int{1, 15, 30})
	viper.Set("candlestick.hours", []int{1, 2, 4, 6, 12})
	viper.Set("candlestick.depths", []int{1, 9, 10, 12, 14, 20, 26, 50})
	viper.Set("candlestick.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
	viper.Set("analysis.storage.limit", 200)

	viper.Set("load.symbols", []string{"BTCUSDT", "ETHUSDT", "TONUSDT"})

	viper.Set("logger.level", "ERROR")
	viper.Set("logger.output_paths", []string{"stdout"})
	viper.Set("logger.error_output_paths", []string{"stdout"})
	viper.Set("logger.stacktrace", false)

	viper.Set("candlestick.logger.level", "ERROR")
	viper.Set("candlestick.logger.output_paths", []string{"./storage/logs/candlestick_output.log"})
	viper.Set("candlestick.logger.error_output_paths", []string{"./storage/logs/candlestick_error.log"})
	viper.Set("candlestick.logger.stacktrace", false)

	viper.Set("analysis.logger.level", "ERROR")
	viper.Set("analysis.logger.output_paths", []string{"./storage/logs/analysis_output.log"})
	viper.Set("analysis.logger.error_output_paths", []string{"./storage/logs/analysis_error.log"})
	viper.Set("analysis.logger.stacktrace", false)

	viper.Set("exchange.logger.level", "ERROR")
	viper.Set("exchange.logger.output_paths", []string{"./storage/logs/exchange_output.log"})
	viper.Set("exchange.logger.error_output_paths", []string{"./storage/logs/exchange_error.log"})
	viper.Set("exchange.logger.stacktrace", false)

	viper.Set("indicator.logger.level", "ERROR")
	viper.Set("indicator.logger.output_paths", []string{"./storage/logs/indicator_output.log"})
	viper.Set("indicator.logger.error_output_paths", []string{"./storage/logs/indicator_error.log"})
	viper.Set("indicator.logger.stacktrace", false)

	viper.Set("price.logger.level", "ERROR")
	viper.Set("price.logger.output_paths", []string{"./storage/logs/price_output.log"})
	viper.Set("price.logger.error_output_paths", []string{"./storage/logs/price_error.log"})
	viper.Set("price.logger.stacktrace", false)

	viper.Set("http_client.logger.level", "ERROR")
	viper.Set("http_client.logger.output_paths", []string{"./storage/logs/http_client_output.log"})
	viper.Set("http_client.logger.error_output_paths", []string{"./storage/logs/http_client_error.log"})
	viper.Set("http_client.logger.stacktrace", false)
}
