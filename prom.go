package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Prepare Prometheus
	prom    = prometheus.NewRegistry()
	metrics = promauto.With(prom)

	// Dishy Info Metrics
	promDishyBootcount = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "bootcount",
		Help:      "Uptime of Dishy",
	}, []string{"id", "hardware_version", "software_version", "country_code"})

	// Dishy Status Metrics
	promDishyGPSValid = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "gps_valid",
		Help:      "Boolean indicator for GPS Valid",
	})
	promDishyGPSSats = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "gps_sats",
		Help:      "Number of available GPS Satellites",
	})
	promDishyMotorsStuck = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "motors_stuck",
		Help:      "Boolean, dishy motors stuck",
	})
	promDishyThermalThrottle = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "thermal_throttle",
		Help:      "Boolean, thermal throttle",
	})
	promDishyThermalShutdown = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "thermal_shutdown",
		Help:      "Boolean, thermal shutdown",
	})
	promDishyMastNotNearVertical = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "mast_not_near_vertical",
		Help:      "Boolean, mast not near vertical",
	})
)

func promInit() {
	// Serve endpoint
	http.Handle("/metrics", promhttp.HandlerFor(prom, promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(promAddr, nil); err != nil {
		log.WithField("Error", err).Fatal("Failed to start Prometheus")
	}
}
