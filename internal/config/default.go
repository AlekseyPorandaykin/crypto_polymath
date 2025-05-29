package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.Set("app.codename", "crypto_polymath")
	viper.Set("app.debug", true)
	viper.Set("http.host", "")
	viper.Set("http.port", "80")
	viper.Set("price.duration.loader", "30s")
	viper.Set("error_mail_to", "alekseip17389@yahoo.com")
	viper.Set("sentry_dsn", "https://b85f50af518fefe30eb6b703f238293b@o4508576056999936.ingest.de.sentry.io/4508576065781840")

	viper.Set("bybit.host", "https://api.bybit.com/")
	viper.Set("binance.spot_host", "https://api.binance.com")
	viper.Set("binance.future_host", "https://fapi.binance.com")
	viper.Set("bitget.host", "https://api.bitget.com/")
	viper.Set("kucoin.host", "https://api.kucoin.com/")
	viper.Set("okx.host", "https://www.okx.com/")
	viper.Set("gateio.host", "https://api.gateio.ws/")
	viper.Set("kraken.host", "https://api.kraken.com/")
	viper.Set("mexc.host", "https://api.mexc.com/")

	//viper.Set("db_connection.driver", "sqlite")
	//viper.Set("db_connection.path_to_db", "./storage/crypto_polymath.db")

	viper.Set("db_connection.driver", "postgres")
	viper.Set("db_connection.username", "crypto_app")
	viper.Set("db_connection.password", "crypto_developer")
	viper.Set("db_connection.host", "37.1.216.169")
	viper.Set("db_connection.port", "5433")
	viper.Set("db_connection.database", "crypto_app")
	viper.Set("db_connection.schema", "crypto_polymath")
	viper.Set("db_connection.max_open_connections", "10")
	viper.Set("db_connection.max_idle_connections", "10")

	viper.Set("candlestick.minutes", []int{1, 15, 30})
	viper.Set("candlestick.hours", []int{1, 2, 4, 6, 12})
	viper.Set("candlestick.depths", []int{1, 8, 9, 10, 12, 14, 20, 26, 50})
	viper.Set("candlestick.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
	viper.Set("analysis.storage.limit", 200)

	viper.Set("load.hot_symbols", []string{"BTCUSDT", "ETHUSDT", "TONUSDT"})
	viper.Set("load.symbols", []string{
		"BTCUSDT",
		"ETHUSDT",
		"TONUSDT",
	})

	viper.Set("logger.level", "INFO")
	viper.Set("logger.logger.alert_level", "ERROR")
	viper.Set("logger.output_paths", []string{"stdout"})
	viper.Set("logger.error_output_paths", []string{"./storage/logs/logger_error.log"})
	viper.Set("logger.stacktrace", false)

	viper.Set("candlestick.logger.level", "INFO")
	viper.Set("candlestick.logger.output_paths", []string{"./storage/logs/candlestick_output.log"})
	viper.Set("candlestick.logger.error_output_paths", []string{"./storage/logs/candlestick_error.log"})
	viper.Set("candlestick.logger.stacktrace", false)

	viper.Set("analysis.logger.level", "INFO")
	viper.Set("analysis.logger.output_paths", []string{"./storage/logs/analysis_output.log"})
	viper.Set("analysis.logger.error_output_paths", []string{"./storage/logs/analysis_error.log"})
	viper.Set("analysis.logger.stacktrace", false)

	viper.Set("exchange.logger.level", "INFO")
	viper.Set("exchange.logger.output_paths", []string{"./storage/logs/exchange_output.log"})
	viper.Set("exchange.logger.error_output_paths", []string{"./storage/logs/exchange_error.log"})
	viper.Set("exchange.logger.stacktrace", false)

	viper.Set("indicator.logger.level", "INFO")
	viper.Set("indicator.logger.output_paths", []string{"./storage/logs/indicator_output.log"})
	viper.Set("indicator.logger.error_output_paths", []string{"./storage/logs/indicator_error.log"})
	viper.Set("indicator.logger.stacktrace", false)

	viper.Set("price.logger.level", "INFO")
	viper.Set("price.logger.output_paths", []string{"./storage/logs/price_output.log"})
	viper.Set("price.logger.error_output_paths", []string{"./storage/logs/price_error.log"})
	viper.Set("price.logger.stacktrace", false)

	viper.Set("http_client.logger.level", "INFO")
	viper.Set("http_client.logger.alert_level", "ERROR")
	viper.Set("http_client.logger.output_paths", []string{"./storage/logs/http_client_output.log", "stderr"})
	viper.Set("http_client.logger.error_output_paths", []string{"./storage/logs/http_client_error.log", "stdout"})
	viper.Set("http_client.logger.stacktrace", false)

	viper.Set("bybit_sender.logger.level", "INFO")
	viper.Set("bybit_sender.logger.alert_level", "WARN")
	viper.Set("bybit_sender.logger.output_paths", []string{"./storage/logs/bybit_sender_output.log"})
	viper.Set("bybit_sender.logger.error_output_paths", []string{"./storage/logs/bybit_sender_error.log", "stdout"})

	viper.Set("loader.logger.level", "DEBUG")
	viper.Set("loader.logger.alert_level", "WARN")
	viper.Set("loader.logger.output_paths", []string{"stdout"})
	viper.Set("loader.logger.error_output_paths", []string{"./storage/logs/loader_error.log"})
	viper.Set("loader.logger.stacktrace", false)

	viper.Set("calculator.logger.level", "INFO")
	viper.Set("calculator.logger.alert_level", "WARN")
	viper.Set("calculator.logger.output_paths", []string{"./storage/logs/calculator_output.log"})
	viper.Set("calculator.logger.error_output_paths", []string{"./storage/logs/calculator_error.log"})
	viper.Set("calculator.logger.stacktrace", false)

	viper.Set("http_server.logger.level", "INFO")
	viper.Set("http_server.logger.alert_level", "WARN")
	viper.Set("http_server.logger.output_paths", []string{"./storage/logs/http_server_output.log", "stdout"})
	viper.Set("http_server.logger.error_output_paths", []string{"./storage/logs/http_server_error.log", "stdout", "stderr"})
	viper.Set("http_server.logger.stacktrace", false)

	viper.Set("indicator_handler.logger.level", "INFO")
	viper.Set("indicator_handler.logger.alert_level", "WARN")
	viper.Set("indicator_handler.logger.output_paths", []string{"stdout"})
	viper.Set("indicator_handler.logger.error_output_paths", []string{"./storage/logs/indicator_handler_error.log"})
	viper.Set("indicator_handler.logger.stacktrace", false)

	viper.Set("rabbit_mq.username", "admin")
	viper.Set("rabbit_mq.password", "crypto_developer_messages")
	viper.Set("rabbit_mq.addr", "37.1.216.169:5672")

	viper.Set("rabbit_mq.exchange", "")
	viper.Set("rabbit_mq.consumer", "candlesticks-queue-consumer")
	viper.Set("rabbit_mq.queue_candlestick", "candlesticks")
	viper.Set("rabbit_mq.queue_action", "actions")
	viper.Set("rabbit_mq.queue_indicator", "indicators")
	viper.Set("rabbit_mq.queue_analytic", "analytics")
	viper.Set("rabbit_mq.queue_candle_indicator", "candle_indicators")
}
