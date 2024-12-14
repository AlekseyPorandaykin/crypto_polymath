package loader

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

const subsystem = "loader"

var calcIndicator = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "calc_indicator",
	Help:      "How much execute indicator calculated.",
}, []string{"exchange", "unit", "interval", "depth"})

var totalCalculatedIndicator = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "calculated_indicator_total",
	Help:      "How much indicator calculated.",
}, []string{"exchange", "unit", "interval", "depth"})

var durationCalcIndicator = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_calc_indicator",
	Help:      "How long execute indicator  calculated in seconds.",
}, []string{"exchange", "unit", "interval", "depth"})

var pricesLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "prices_loaded",
	Help:      "How much prices loaded from external exchange.",
}, []string{"exchange"})

var candlestickLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "candlestick_loaded",
	Help:      "How much execute candlestick loader from external exchange.",
}, []string{"exchange", "interval"})

var candlestickLoadedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "candlestick_loaded_total",
	Help:      "How much candlesticks loaded from external exchange.",
}, []string{"exchange", "unit"})

var futureCandlestickLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "future_candlestick_loaded",
	Help:      "How much execute future candlestick loader from external exchange.",
}, []string{"exchange", "interval"})

var futureCandlestickLoadedDuration = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "future_candlestick_loaded_seconds",
	Help:      "How long future candlesticks loaded in seconds.",
}, []string{"exchange", "unit"})

var countsPricesLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "count_prices_loaded",
	Help:      "How much loaded from external exchange.",
}, []string{"exchange"})

var durationPricesLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_prices_loaded",
	Help:      "How long prices loaded from external exchange in seconds.",
}, []string{"exchange"})

var durationCandlestickLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_candlestick_loaded_milliseconds",
	Help:      "How long candlestick loaded from external exchange in ms.",
}, []string{"exchange", "interval"})

var deleteCandlestick = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "delete_candlestick",
	Help:      "How much candlesticks deleted.",
})

var durationDeleteCandlestick = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_delete_candlestick",
	Help:      "How long candlesticks deleted in seconds.",
})

var exchangeSymbolTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "exchange_symbol_loaded_total",
	Help:      "How much symbol loaded.",
}, []string{"exchange"})

var exchangeSymbolDuration = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "exchange_symbol_loaded_seconds",
	Help:      "How long symbol loaded in seconds.",
}, []string{"exchange"})

var errorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "error_total",
	Help:      "How much errors.",
}, []string{"exchange", "reason"})

func durationPricesLoadedHelper(exchange string) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		countsPricesLoaded.WithLabelValues(exchange).Inc()
		durationPricesLoaded.WithLabelValues(exchange).Add(duration.Seconds())
	}
}
func durationCandlestickLoadedHelper(exchange, unit string, interval int) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		candlestickLoaded.WithLabelValues(exchange, unit).Inc()
		durationCandlestickLoaded.WithLabelValues(exchange, unit).Add(float64(duration.Milliseconds()))
	}
}

func durationFutureCandlestickLoadedHelper(exchange, unit string) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		futureCandlestickLoadedDuration.WithLabelValues(exchange, unit).Add(float64(duration.Milliseconds()))
	}
}
func ExchangeSymbolLoadedHelper(exchange string) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		exchangeSymbolTotal.WithLabelValues(exchange).Inc()
		exchangeSymbolDuration.WithLabelValues(exchange).Add(duration.Seconds())
	}
}

func deleteIndicatorHelper() func() {
	start := time.Now()

	return func() {
		deleteCandlestick.Inc()
		durationDeleteCandlestick.Add(time.Since(start).Seconds())
	}
}
func calcIndicatorHelper(exchangeName string, unit string, interval, depth int) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		calcIndicator.WithLabelValues(exchangeName, unit, strconv.Itoa(interval), strconv.Itoa(depth)).Inc()
		durationCalcIndicator.WithLabelValues(exchangeName, unit, strconv.Itoa(interval), strconv.Itoa(depth)).Add(duration.Seconds())
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(pricesLoaded, durationPricesLoaded, countsPricesLoaded)
	prometheus.DefaultRegisterer.MustRegister(candlestickLoaded, durationCandlestickLoaded, candlestickLoadedTotal)
	prometheus.DefaultRegisterer.MustRegister(deleteCandlestick, durationDeleteCandlestick)
	prometheus.DefaultRegisterer.MustRegister(exchangeSymbolTotal, exchangeSymbolDuration)
	prometheus.DefaultRegisterer.MustRegister(calcIndicator, totalCalculatedIndicator, durationCalcIndicator)
	prometheus.DefaultRegisterer.MustRegister(futureCandlestickLoaded, futureCandlestickLoadedDuration)
}
