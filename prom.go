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

	// InternalMetrics
	promDishyGRPCTime = metrics.NewHistogram(prometheus.HistogramOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "grpc_time",
		Help:      "Time spend interacting with Dishy GRPC",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
	})
	promDishyUpdateTime = metrics.NewHistogram(prometheus.HistogramOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "update_time",
		Help:      "Total time to pull all data from Dishy",
		Buckets:   prometheus.ExponentialBuckets(1, 3, 8),
	})
	promDishyUpdates = metrics.NewCounter(prometheus.CounterOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "updates",
		Help:      "Number of Dishy updates",
	})

	// Dishy Info Metrics
	promDishyBootcount = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "bootcount",
		Help:      "Uptime of Dishy",
	}, []string{"id", "hardware_version", "software_version", "manufactured_version", "country_code"})

	// Dishy GPS Metrics
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

	// Alerts
	promDishyAlertMetrics map[string]*prometheus.Gauge // Accessor for alerts
	promDishyMotorsStuck  = metrics.NewGauge(prometheus.GaugeOpts{
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
	promDishyRoaming = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "roaming",
		Help:      "Boolean, dishy is roaming",
	})
	promDishyUnexpectedLocation = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "unexpected_location",
		Help:      "Boolean, dishy is in an unexpected location",
	})
	promDishySlowEthernet = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "slow_ethernet_speeds",
		Help:      "Boolean, slow ethernet speeds",
	})
	promDishyObstructed = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "obstructed",
		Help:      "Boolean, dishy is obstructed",
	})
	promDishyFractionObstructed = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "fraction_obstructed",
		Help:      "Boolean, dishy is obstructed",
	})
	promDishyAvgObstructedDurationS = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "avg_prolonged_obstruction_duration_s",
		Help:      "Boolean, average duration of obstruction",
	})
	promDishyOutage = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "outage",
		Help:      "Boolean, current Dishy outage status",
	})
	promDishyAlerts = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "alerts",
		Help:      "Number of current alerts",
	})
	promDishyPopPingDropRate = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "pop_ping_drop_rate",
		Help:      "Current pop ping drop rate",
	})
	promDishyPopPingLatencyMs = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "pop_ping_latency_ms",
		Help:      "Current pop ping latency",
	})
	promDishyDLTputBps = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "downlink_throughput_bps",
		Help:      "Current downlink throughput in bits persecond",
	})
	promDishyULTputBps = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "uplink_throughput_bps",
		Help:      "Current uplink throughput in bits persecond",
	})
	promDishyAzimuthDeg = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "boresight_azimuth_deg",
		Help:      "Boresight azimum in degrees",
	})
	promDishyElevationDeg = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "boresight_elevation_deg",
		Help:      "Boresight elevation in degrees",
	})
	promDishyEthSpeedMbps = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "eth_speed_mbps",
		Help:      "Ethernet speed in mbps",
	})

	// Outage Metrics
	promDishyOutages = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "outages",
		Help:      "Number of recent outages",
	}, []string{"cause"})
	promDishyAvgOutageDuration = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "outage_duration_sec_avg",
		Help:      "Avg Outage Duration",
	}, []string{"cause"})
	promDishySumOutageDuration = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "outage_duration_sec_sum",
		Help:      "Total Outage Duration",
	}, []string{"cause"})
)

func promInit() {
	// Add Alerts to Alert Accessor
	// TODO Consider a single metric with alert name as label
	promDishyAlertMetrics = make(map[string]*prometheus.Gauge)
	promDishyAlertMetrics["MotorsStuck"] = &promDishyMotorsStuck
	promDishyAlertMetrics["ThermalThrottle"] = &promDishyThermalThrottle
	promDishyAlertMetrics["ThermalShutdown"] = &promDishyThermalShutdown
	promDishyAlertMetrics["MastNotNearVertical"] = &promDishyMastNotNearVertical
	promDishyAlertMetrics["UnexpectedLocation"] = &promDishyUnexpectedLocation
	promDishyAlertMetrics["SlowEthernetSpeeds"] = &promDishySlowEthernet
	promDishyAlertMetrics["Roaming"] = &promDishyRoaming
	// Serve endpoint
	http.Handle("/metrics", promhttp.HandlerFor(prom, promhttp.HandlerOpts{}))
	log.WithField("Listen", promAddr).Info("Prometheus Starting")
	if err := http.ListenAndServe(promAddr, nil); err != nil {
		log.WithField("Error", err).Fatal("Failed to start Prometheus")
	}
}
