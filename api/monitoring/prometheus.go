package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusEndpoint struct {
	PostRequest chan bool
	GetRequest  chan bool
	endpoint    string
	addr        string
}

func NewPrometheusEndpoint(endpoint, addr string) *PrometheusEndpoint {
	prch := make(chan bool)
	grch := make(chan bool)
	return &PrometheusEndpoint{
		PostRequest: prch,
		GetRequest:  grch,
		endpoint:    endpoint,
		addr:        addr,
	}
}

func (p *PrometheusEndpoint) RunPrometheusEndpoint() error {
	p.recordMetrics()
	http.Handle(p.endpoint, promhttp.Handler())
	return http.ListenAndServe(p.addr, nil)
}

func (p *PrometheusEndpoint) recordMetrics() {
	go func() {
		for {
			select {
			case <-p.GetRequest:
				getRequestCounter.Inc()
			case <-p.PostRequest:
				postRequestCounter.Inc()
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

var (
	getRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "secret_server_get_request_total",
		Help: "The total number of successfully finished get requests",
	})
	postRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "secret_server_post_request_total",
		Help: "The total number of successfully finished post requests",
	})
)
