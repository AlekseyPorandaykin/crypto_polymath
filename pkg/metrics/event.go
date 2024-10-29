package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespaceEvent = "event"

var EventCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceEvent,
	Name:      "count_event",
	Help:      "How much execute events.",
}, []string{"event_name", "step"})

var EventDurationQuery = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceEvent,
	Name:      "duration_query",
	Help:      "How long event handled(in milliseconds).",
}, []string{"event_name"})

func init() {
	prometheus.DefaultRegisterer.MustRegister(EventCount, EventDurationQuery)
}
