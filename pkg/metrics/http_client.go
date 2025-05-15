package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

var namespaceHTTPClient = "http_client"

var httpClientCountRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceHTTPClient,
	Name:      "count_request",
	Help:      "How much requests executed.",
}, []string{"host", "method", "path"})

var httpClientCountErrorRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceHTTPClient,
	Name:      "count_error_request",
	Help:      "How much error requests.",
}, []string{"host", "method", "path"})

var httpClientCountResponseCode = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceHTTPClient,
	Name:      "count_response_code",
	Help:      "Response codes.",
}, []string{"host", "method", "path", "status_code"})

var httpClientDurationRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespaceHTTPClient,
	Name:      "duration_request",
	Help:      "How lon requests executed(in milliseconds).",
}, []string{"host", "method", "path"})

func HTTPClientQueryHelper(host, method, path string) func() {
	now := time.Now()
	return func() {
		httpClientCountRequest.WithLabelValues(host, method, path).Inc()
		httpClientDurationRequest.WithLabelValues(host, method, path).Add(float64(time.Since(now).Milliseconds()))
	}
}

func init() {
	prometheus.DefaultRegisterer.MustRegister(
		httpClientCountRequest,
		httpClientCountErrorRequest,
		httpClientCountResponseCode,
		httpClientDurationRequest,
	)
}

type HTTPSender interface {
	Do(r *http.Request) (*http.Response, error)
}

type HTTPSenderWithMetrics struct {
	sender HTTPSender
}

func NewHTTPSenderWithMetrics(sender HTTPSender) HTTPSender {
	return &HTTPSenderWithMetrics{sender: sender}
}

func (s *HTTPSenderWithMetrics) Do(r *http.Request) (*http.Response, error) {
	defer HTTPClientQueryHelper(r.URL.Host, r.Method, r.URL.Path)()
	resp, err := s.sender.Do(r)
	if resp != nil {
		httpClientCountResponseCode.WithLabelValues(r.URL.Host, r.Method, r.URL.Path, strconv.Itoa(resp.StatusCode)).Inc()
	}
	if err != nil {
		httpClientCountErrorRequest.WithLabelValues(r.URL.Host, r.Method, r.URL.Path).Inc()
	}
	return resp, err
}
