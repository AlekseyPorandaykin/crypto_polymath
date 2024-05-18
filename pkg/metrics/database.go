package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var namespaceDB = "db"

var countQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceDB,
	Name:      "count_query",
	Help:      "How much queries executed.",
}, []string{"db", "query"})

var durationQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceDB,
	Name:      "duration_query",
	Help:      "How lon queries executed(in seconds).",
}, []string{"db", "query"})

func DBQueryHelper(db, query string) func() {
	now := time.Now()
	return func() {
		countQuery.WithLabelValues(db, query).Inc()
		durationQuery.WithLabelValues(db, query).Add(time.Since(now).Seconds())
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(countQuery, durationQuery)
}
