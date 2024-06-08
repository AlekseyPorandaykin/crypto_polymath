package calculator

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

const subsystem = "calculator"

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

var deleteIndicator = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "delete_indicator",
	Help:      "How much indicator deleted.",
})

var durationDeleteIndicator = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_delete_indicator",
	Help:      "How long indicator deleted in seconds.",
})
var deleteAnalysis = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "delete_analysis",
	Help:      "How much analysis deleted.",
})

var durationDeleteAnalysis = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "duration_delete_analysis",
	Help:      "How long analysis deleted in seconds.",
})
var errorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: viper.GetString("app.codename"),
	Subsystem: subsystem,
	Name:      "error_total",
	Help:      "How much errors.",
}, []string{"exchange", "reason"})

func calcIndicatorHelper(exchangeName string, unit string, interval, depth int) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		calcIndicator.WithLabelValues(exchangeName, unit, strconv.Itoa(interval), strconv.Itoa(depth)).Inc()
		durationCalcIndicator.WithLabelValues(exchangeName, unit, strconv.Itoa(interval), strconv.Itoa(depth)).Add(duration.Seconds())
	}
}

func deleteIndicatorHelper() func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		deleteIndicator.Inc()
		durationDeleteIndicator.Add(duration.Seconds())
	}
}
func deleteAnalysisHelper() func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		deleteAnalysis.Inc()
		durationDeleteAnalysis.Add(duration.Seconds())
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(calcIndicator)
	prometheus.DefaultRegisterer.MustRegister(durationCalcIndicator)
	prometheus.DefaultRegisterer.MustRegister(deleteIndicator)
	prometheus.DefaultRegisterer.MustRegister(durationDeleteIndicator)
	prometheus.DefaultRegisterer.MustRegister(totalCalculatedIndicator)
	prometheus.DefaultRegisterer.MustRegister(errorTotal)
	prometheus.DefaultRegisterer.MustRegister(deleteAnalysis, durationDeleteAnalysis)
}
