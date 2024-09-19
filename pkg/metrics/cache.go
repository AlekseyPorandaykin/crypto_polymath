package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var namespaceCache = "cache"

var cacheCountQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceCache,
	Name:      "count_query",
	Help:      "How much queries executed.",
}, []string{"db", "query"})

var cacheDurationQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceCache,
	Name:      "duration_query",
	Help:      "How lon queries executed(in milliseconds).",
}, []string{"db", "query"})

func CacheQueryHelper(db, query string) func() {
	now := time.Now()
	return func() {
		cacheCountQuery.WithLabelValues(db, query).Inc()
		cacheDurationQuery.WithLabelValues(db, query).Add(float64(time.Since(now).Milliseconds()))
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(cacheCountQuery, cacheDurationQuery)
}
