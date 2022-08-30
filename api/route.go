package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var H http.Handler

func InitRoute(gatherers prometheus.Gatherers) {
	H = promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			//ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		})

	http.HandleFunc("/metrics", MetricsAPI)
	http.HandleFunc("/daily/yersterday", Daily)
}
