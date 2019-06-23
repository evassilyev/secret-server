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
	getRequestResponseSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "get_request_response_time_sm",
		Help:       "Get requests responce time summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.004, 0.99: 0.001},
	})
	postRequestResponseSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "post_request_response_time_sm",
		Help:       "Post requests responce time summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.004, 0.99: 0.001},
	})
)

type PrometheusEndpoint struct {
	PostRequest   chan bool
	GetRequest    chan bool
	PostRequestRT chan time.Duration
	GetRequestRT  chan time.Duration
	endpoint      string
	addr          string
}

func NewPrometheusEndpoint(endpoint, addr string) *PrometheusEndpoint {
	prch := make(chan bool)
	grch := make(chan bool)
	prrtch := make(chan time.Duration)
	grrtch := make(chan time.Duration)
	return &PrometheusEndpoint{
		PostRequest:   prch,
		GetRequest:    grch,
		endpoint:      endpoint,
		addr:          addr,
		PostRequestRT: prrtch,
		GetRequestRT:  grrtch,
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
				getRequestResponseSummary.Observe(float64(v))
			case v = <-p.PostRequestRT:
				postRequestResponseSummary.Observe(float64(v))
			default:
				//time.Sleep(2 * time.Second)
			}
		}
	}()
}
