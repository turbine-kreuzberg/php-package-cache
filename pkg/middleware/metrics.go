package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	inFlight          prometheus.Gauge
	requestSize       *prometheus.HistogramVec
	responseSize      *prometheus.HistogramVec
)

func InitMetrics(gitHash, gitRef string) {
	r := prometheus.DefaultRegisterer

	version := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"git_hash": gitHash,
			"git_ref":  gitRef,
		},
	})
	version.Set(1)
	r.MustRegister(version)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		},
		[]string{"code", "method"},
	)
	r.MustRegister(httpRequestsTotal)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of latencies for HTTP requests.",
			//			Buckets: []float64{.05, 0.1, .25, .5, .75, 1, 2, 5, 20, 60},
		},
		[]string{"code", "method"},
	)
	r.MustRegister(requestDuration)

	inFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of requests being served.",
		},
	)
	r.MustRegister(inFlight)

	requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Histogram of HTTP request size.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"code", "method"},
	)
	r.MustRegister(requestSize)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Histogram of response size for HTTP requests.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"code", "method"},
		//		[]string{"code", "method", "handler"},
	)
	r.MustRegister(responseSize)
}

func InstrumentHttpHandler(next http.Handler) http.Handler {
	next = promhttp.InstrumentHandlerCounter(httpRequestsTotal, next)
	next = promhttp.InstrumentHandlerDuration(requestDuration, next)
	next = promhttp.InstrumentHandlerInFlight(inFlight, next)
	next = promhttp.InstrumentHandlerRequestSize(requestSize, next)
	next = promhttp.InstrumentHandlerResponseSize(responseSize, next)
	//next = promhttp.InstrumentHandlerResponseSize(
	//	responseSize.MustCurryWith(prometheus.Labels{"handler": handlerName}), next)
	//	next = promhttp.InstrumentHandlerTimeToWriteHeader(httpRequestsTotal, next)

	return next
}
