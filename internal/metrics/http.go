package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type HTTPMetrics interface {
	RequestDuration(method string, route string, duration time.Duration)
	RequestTotal(method string, route string, status int)
}

type httpMetrics struct {
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec
}

func NewHTTPMetrics(reg prometheus.Registerer) *httpMetrics {
	const subsystem = "http"

	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	requestTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total HTTP requests",
		}, []string{"method", "route", "status"},
	)

	return &httpMetrics{
		requestDuration: requestDuration,
		requestTotal:    requestTotal,
	}
}

func (h *httpMetrics) RequestDuration(method string, route string, duration time.Duration) {
	h.requestDuration.WithLabelValues(method, route).Observe(duration.Seconds())
}

func (h *httpMetrics) RequestTotal(method string, route string, status int) {
	h.requestTotal.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
}
