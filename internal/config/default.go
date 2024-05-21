package config

import (
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
	viper.Set("db_connection.path_to_db", "./storage/crypto_polymath.db")

	viper.Set("candlestick.minutes", []int{1, 15, 30})
	viper.Set("candlestick.hours", []int{1, 2, 4, 6, 12})
	viper.Set("candlestick.depths", []int{1, 10, 20, 50})
	viper.Set("candlestick.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
	viper.Set("indicator.storage.limit", 200)
}
