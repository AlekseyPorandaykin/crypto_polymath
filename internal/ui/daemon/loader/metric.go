package loader

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

const subsystem = "loader"

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
	Help:      "How much candlestick loaded from external exchange.",
}, []string{"exchange", "interval"})

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
	Name:      "duration_candlestick_loaded",
	Help:      "How long candlestick loaded from external exchange in seconds.",
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

func durationPricesLoadedHelper(exchange string) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		zap.L().Debug("load prices", zap.String("exchange", exchange), zap.String("duration", duration.String()))
		countsPricesLoaded.WithLabelValues(exchange).Inc()
		durationPricesLoaded.WithLabelValues(exchange).Add(duration.Seconds())
	}
}
func durationCandlestickLoadedHelper(exchange, unit string, interval int) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		zap.L().Debug(
			"load candlesticks",
			zap.String("unit", unit),
			zap.Int("interval", interval),
			zap.String("duration", duration.String()),
		)
		candlestickLoaded.WithLabelValues(exchange, unit).Inc()
		durationCandlestickLoaded.WithLabelValues(exchange, unit).Add(duration.Seconds())
	}
}

func deleteIndicatorHelper() func() {
	start := time.Now()

	return func() {
		deleteCandlestick.Inc()
		durationDeleteCandlestick.Add(time.Since(start).Seconds())
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(pricesLoaded)
	prometheus.DefaultRegisterer.MustRegister(candlestickLoaded)
	prometheus.DefaultRegisterer.MustRegister(durationPricesLoaded)
	prometheus.DefaultRegisterer.MustRegister(durationCandlestickLoaded)
	prometheus.DefaultRegisterer.MustRegister(countsPricesLoaded)
	prometheus.DefaultRegisterer.MustRegister(deleteCandlestick)
	prometheus.DefaultRegisterer.MustRegister(durationDeleteCandlestick)
}
