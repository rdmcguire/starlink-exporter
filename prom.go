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
	promDishyFailures = metrics.NewCounter(prometheus.CounterOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "failures",
		Help:      "Number of Dishy request failures",
	})
	promDishyRequests = metrics.NewCounter(prometheus.CounterOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "requests",
		Help:      "Number of Dishy requests",
	})
	promDishyFailing = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "exporter",
		Name:      "failing",
		Help:      "Boolean indicator if requests to Dishy are failing",
	})

	// Dishy Info Metrics
	promDishyBootcount = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "bootcount",
		Help:      "Dishy Boot Count",
	}, []string{"id", "hardware_version", "software_version", "manufactured_version", "country_code"})
	promDishyUptimeS = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "uptime_s",
		Help:      "Uptime of Dishy",
	})

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
	promDishyAlertStatus = metrics.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "alert_status",
		Help:      "Status of Alerts",
	}, []string{"alert"})
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
		Help:      "Percent Obstructed",
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
	promDishyOutageHistogram = metrics.NewHistogram(prometheus.HistogramOpts{
		Namespace: "starlink",
		Subsystem: "dishy",
		Name:      "outage_times",
		Help:      "Histogram of outage times",
		Buckets:   prometheus.ExponentialBucketsRange(0.25, 3600, 15),
	})
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
	// Serve endpoint
	http.Handle("/metrics", promhttp.HandlerFor(prom, promhttp.HandlerOpts{}))
	log.WithField("Listen", promAddr).Info("Prometheus Starting")
	if err := http.ListenAndServe(promAddr, nil); err != nil {
		log.WithField("Error", err).Fatal("Failed to start Prometheus")
	}
}
