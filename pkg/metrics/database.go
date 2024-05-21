package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var namespaceDB = "db"

var dbCountQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceDB,
	Name:      "count_query",
	Help:      "How much queries executed.",
}, []string{"db", "query"})

var dbDurationQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceDB,
	Name:      "duration_query",
	Help:      "How lon queries executed(in seconds).",
}, []string{"db", "query"})

func DBQueryHelper(db, query string) func() {
	now := time.Now()
	return func() {
		dbCountQuery.WithLabelValues(db, query).Inc()
		dbDurationQuery.WithLabelValues(db, query).Add(time.Since(now).Seconds())
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(dbCountQuery, dbDurationQuery)
}
