package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// StartServer starts the Prometheus HTTP server
func StartServer(addr string, metrics *Metrics, logger *logrus.Logger) error {
	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	logger.WithField("Listen", addr).Info("Prometheus Starting")
	return http.ListenAndServe(addr, nil)
}
