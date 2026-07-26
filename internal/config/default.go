package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("metric.port", "8080")
	viper.SetDefault("app.codename", "crypto_polymath")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "80")
	viper.SetDefault("grpc.port", "50052")
	viper.SetDefault("price.duration.loader", "30s")
	viper.SetDefault("error_mail_to", "alekseip17389@yahoo.com")
	viper.SetDefault("sentry_dsn", "https://b85f50af518fefe30eb6b703f238293b@o4508576056999936.ingest.de.sentry.io/4508576065781840")

	viper.SetDefault("bybit.host", "https://api.bybit.com/")
	viper.SetDefault("binance.spot_host", "https://api.binance.com")
	viper.SetDefault("binance.future_host", "https://fapi.binance.com")
	viper.SetDefault("bitget.host", "https://api.bitget.com/")
	viper.SetDefault("kucoin.host", "https://api.kucoin.com/")
	viper.SetDefault("okx.host", "https://www.okx.com/")
	viper.SetDefault("gateio.host", "https://api.gateio.ws/")
	viper.SetDefault("kraken.host", "https://api.kraken.com/")
	viper.SetDefault("mexc.host", "https://api.mexc.com/")

	//viper.SetDefault("db_connection.driver", "sqlite")
	//viper.SetDefault("db_connection.path_to_db", "./storage/crypto_polymath.db")

	viper.SetDefault("db_connection.driver", "postgres")
	viper.SetDefault("db_connection.username", "crypto_app")
	viper.SetDefault("db_connection.password", "crypto_developer")
	viper.SetDefault("db_connection.host", "127.0.0.1")
	viper.SetDefault("db_connection.port", "5433")
	viper.SetDefault("db_connection.database", "crypto_app")
	viper.SetDefault("db_connection.schema", "crypto_polymath")
	viper.SetDefault("db_connection.max_open_connections", 10)
	viper.SetDefault("db_connection.max_idle_connections", 10)

	viper.SetDefault("candlestick.minutes", []int{1, 15, 30})
	viper.SetDefault("candlestick.hours", []int{1, 2, 4, 6, 12})
	viper.SetDefault("candlestick.depths", []int{1, 8, 9, 10, 12, 14, 20, 26, 50})
	viper.SetDefault("candlestick.storage.limit", 5000)
	viper.SetDefault("indicator.storage.limit", 5000)
	viper.SetDefault("analysis.storage.limit", 5000)

	viper.SetDefault("load.hot_symbols", []string{"BTCUSDT", "ETHUSDT", "TONUSDT"})
	viper.SetDefault("load.symbols", []string{
		"BTCUSDT",
		"ETHUSDT",
		"TONUSDT",
	})

	viper.SetDefault("logger.level", "INFO")
	viper.SetDefault("logger.alert_level", "ERROR")
	viper.SetDefault("logger.output_paths", []string{"stdout"})
	viper.SetDefault("logger.error_output_paths", []string{"./storage/logs/logger_error.log", "stdout"})
	viper.SetDefault("logger.stacktrace", false)

	viper.SetDefault("repository.logger.level", "INFO")
	viper.SetDefault("repository.logger.alert_level", "ERROR")
	viper.SetDefault("repository.logger.output_paths", []string{"stdout", "./storage/logs/repository_output.log"})
	viper.SetDefault("repository.logger.error_output_paths", []string{"./storage/logs/repository_error.log", "stdout"})
	viper.SetDefault("repository.logger.stacktrace", false)

	viper.SetDefault("candlestick.logger.level", "INFO")
	viper.SetDefault("candlestick.logger.output_paths", []string{"./storage/logs/candlestick_output.log", "stdout"})
	viper.SetDefault("candlestick.logger.error_output_paths", []string{"./storage/logs/candlestick_error.log", "stdout"})
	viper.SetDefault("candlestick.logger.stacktrace", false)

	viper.SetDefault("analysis.logger.level", "INFO")
	viper.SetDefault("analysis.logger.output_paths", []string{"./storage/logs/analysis_output.log", "stdout"})
	viper.SetDefault("analysis.logger.error_output_paths", []string{"./storage/logs/analysis_error.log", "stdout"})
	viper.SetDefault("analysis.logger.stacktrace", false)

	viper.SetDefault("exchange.logger.level", "INFO")
	viper.SetDefault("exchange.logger.output_paths", []string{"./storage/logs/exchange_output.log", "stdout"})
	viper.SetDefault("exchange.logger.error_output_paths", []string{"./storage/logs/exchange_error.log", "stdout"})
	viper.SetDefault("exchange.logger.stacktrace", false)

	viper.SetDefault("indicator.logger.level", "INFO")
	viper.SetDefault("indicator.logger.output_paths", []string{"./storage/logs/indicator_output.log", "stdout"})
	viper.SetDefault("indicator.logger.error_output_paths", []string{"./storage/logs/indicator_error.log", "stdout"})
	viper.SetDefault("indicator.logger.stacktrace", false)

	viper.SetDefault("price.logger.level", "INFO")
	viper.SetDefault("price.logger.output_paths", []string{"./storage/logs/price_output.log", "stdout"})
	viper.SetDefault("price.logger.error_output_paths", []string{"./storage/logs/price_error.log", "stdout"})
	viper.SetDefault("price.logger.stacktrace", false)

	viper.SetDefault("http_client.logger.level", "INFO")
	viper.SetDefault("http_client.logger.alert_level", "ERROR")
	viper.SetDefault("http_client.logger.output_paths", []string{"./storage/logs/http_client_output.log", "stdout"})
	viper.SetDefault("http_client.logger.error_output_paths", []string{"./storage/logs/http_client_error.log", "stdout"})
	viper.SetDefault("http_client.logger.stacktrace", false)

	viper.SetDefault("bybit_sender.logger.level", "INFO")
	viper.SetDefault("bybit_sender.logger.alert_level", "WARN")
	viper.SetDefault("bybit_sender.logger.output_paths", []string{"./storage/logs/bybit_sender_output.log", "stdout"})
	viper.SetDefault("bybit_sender.logger.error_output_paths", []string{"./storage/logs/bybit_sender_error.log", "stdout"})

	viper.SetDefault("loader.logger.level", "INFO")
	viper.SetDefault("loader.logger.alert_level", "WARN")
	viper.SetDefault("loader.logger.output_paths", []string{"stdout"})
	viper.SetDefault("loader.logger.error_output_paths", []string{"./storage/logs/loader_error.log"})
	viper.SetDefault("loader.logger.stacktrace", false)

	viper.SetDefault("calculator.logger.level", "INFO")
	viper.SetDefault("calculator.logger.alert_level", "WARN")
	viper.SetDefault("calculator.logger.output_paths", []string{"./storage/logs/calculator_output.log", "stdout"})
	viper.SetDefault("calculator.logger.error_output_paths", []string{"./storage/logs/calculator_error.log"})
	viper.SetDefault("calculator.logger.stacktrace", false)

	viper.SetDefault("http_server.logger.level", "INFO")
	viper.SetDefault("http_server.logger.alert_level", "WARN")
	viper.SetDefault("http_server.logger.output_paths", []string{"./storage/logs/http_server_output.log", "stdout"})
	viper.SetDefault("http_server.logger.error_output_paths", []string{"./storage/logs/http_server_error.log", "stdout", "stderr"})
	viper.SetDefault("http_server.logger.stacktrace", false)

	viper.SetDefault("indicator_handler.logger.level", "INFO")
	viper.SetDefault("indicator_handler.logger.alert_level", "WARN")
	viper.SetDefault("indicator_handler.logger.output_paths", []string{"stdout"})
	viper.SetDefault("indicator_handler.logger.error_output_paths", []string{"./storage/logs/indicator_handler_error.log"})
	viper.SetDefault("indicator_handler.logger.stacktrace", false)

	viper.SetDefault("rabbit_mq.username", "admin")
	viper.SetDefault("rabbit_mq.password", "crypto_developer_messages")
	viper.SetDefault("rabbit_mq.addr", "127.0.0.1:5672")

	viper.SetDefault("events.queue_candlestick", "candlesticks")
	viper.SetDefault("events.queue_action", "actions")
	viper.SetDefault("events.queue_indicator", "indicators")
	viper.SetDefault("events.queue_analytic", "analytics")
	viper.SetDefault("events.queue_candle_indicator", "candle_indicators")
	viper.SetDefault("events.queue_batch_size", 2000)
}
