package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics for collection
var (
	getRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "secret_server_get_request_total",
		Help: "The total number of get requests",
	})
	postRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "secret_server_post_request_total",
		Help: "The total number of post requests",
	})
	getRequestResponseHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "get_request_response_time",
		Help:    "Get requests responce time histogram",
		Buckets: []float64{0.5, 1, 2, 3, 5, 10},
	})
	postRequestResponseHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_request_response_time",
		Help:    "Post requests responce time histogram",
		Buckets: []float64{0.5, 1, 2, 3, 5, 10},
	})
)

// TODO remove unused
type PrometheusEndpoint struct {
	PostRequest                    chan bool
	GetRequest                     chan bool
	PostRequestRT                  chan time.Duration
	GetRequestRT                   chan time.Duration
	GetRequestRTMetrics            chan prometheus.Metric
	PostRequestRTMetrics           chan prometheus.Metric
	endpoint                       string
	addr                           string
	PostRequestsResponseTimeBuffer *BufferStorage
	GetRequestsResponseTimeBuffer  *BufferStorage
}

const responce_time_buffer_size = 1000

func NewPrometheusEndpoint(endpoint, addr string) *PrometheusEndpoint {
	prch := make(chan bool)
	grch := make(chan bool)
	prrtch := make(chan time.Duration)
	grrtch := make(chan time.Duration)
	prrtmch := make(chan prometheus.Metric)
	grrtmch := make(chan prometheus.Metric)
	prtb := NewBufferStorage(responce_time_buffer_size)
	grtb := NewBufferStorage(responce_time_buffer_size)
	return &PrometheusEndpoint{
		PostRequest:                    prch,
		GetRequest:                     grch,
		endpoint:                       endpoint,
		addr:                           addr,
		PostRequestsResponseTimeBuffer: prtb,
		GetRequestsResponseTimeBuffer:  grtb,
		PostRequestRT:                  prrtch,
		GetRequestRT:                   grrtch,
		PostRequestRTMetrics:           prrtmch,
		GetRequestRTMetrics:            grrtmch,
	}
}

func (p *PrometheusEndpoint) RunPrometheusEndpoint() {
	p.recordMetrics()
	http.Handle(p.endpoint, promhttp.Handler())
	http.ListenAndServe(p.addr, nil)
}

func (p *PrometheusEndpoint) recordMetrics() {
	go func() {
		for {
			var v time.Duration
			select {
			case <-p.GetRequest:
				getRequestCounter.Inc()
			case <-p.PostRequest:
				postRequestCounter.Inc()
			case v = <-p.GetRequestRT:
				getRequestResponseHistogram.Observe(float64(v))
			case v = <-p.PostRequestRT:
				postRequestResponseHistogram.Observe(float64(v))
			default:
				//time.Sleep(2 * time.Second)
			}
		}
	}()
}
