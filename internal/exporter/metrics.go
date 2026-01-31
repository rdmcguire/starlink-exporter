package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the exporter
type Metrics struct {
	Registry *prometheus.Registry

	// Internal exporter metrics
	GRPCTime    prometheus.Histogram
	UpdateTime  prometheus.Histogram
	Updates     prometheus.Counter
	Failures    prometheus.Counter
	Requests    prometheus.Counter
	Failing     prometheus.Gauge

	// Dishy info metrics
	Bootcount *prometheus.GaugeVec
	UptimeS   prometheus.Gauge

	// Dishy GPS metrics
	GPSValid prometheus.Gauge
	GPSSats  prometheus.Gauge

	// Dishy alert metrics
	AlertStatus         *prometheus.GaugeVec
	Alerts              prometheus.Gauge
	Obstructed          prometheus.Gauge
	FractionObstructed  prometheus.Gauge
	AvgObstructedDurS   prometheus.Gauge
	Outage              prometheus.Gauge

	// Dishy throughput/latency metrics
	PopPingDropRate   prometheus.Gauge
	PopPingLatencyMs  prometheus.Gauge
	DLThroughputBps   prometheus.Gauge
	ULThroughputBps   prometheus.Gauge

	// Dishy orientation metrics
	AzimuthDeg    prometheus.Gauge
	ElevationDeg  prometheus.Gauge
	EthSpeedMbps  prometheus.Gauge

	// Dishy outage metrics
	OutageHistogram   prometheus.Histogram
	Outages           *prometheus.GaugeVec
	AvgOutageDuration *prometheus.GaugeVec
	SumOutageDuration *prometheus.GaugeVec

	// NEW: Additional status metrics
	StowRequested            prometheus.Gauge
	SNRAboveNoiseFloor       prometheus.Gauge
	SNRPersistentlyLow       prometheus.Gauge
	SoftwareUpdateState      *prometheus.GaugeVec
	SoftwareUpdateProgress   prometheus.Gauge
	TiltAngleDeg             prometheus.Gauge
	AttitudeUncertaintyDeg   prometheus.Gauge
	SecondsToFirstNonempty   prometheus.Gauge
	IsCellDisabled           prometheus.Gauge
	MobilityClass            *prometheus.GaugeVec

	// NEW: Additional obstruction stats
	TimeObstructedS       prometheus.Gauge
	ObstructionIntervalS  prometheus.Gauge
	ObstructionValidS     prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	factory := promauto.With(registry)

	m := &Metrics{
		Registry: registry,

		// Internal exporter metrics
		GRPCTime: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "grpc_time",
			Help:      "Time spent interacting with Dishy GRPC",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
		}),
		UpdateTime: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "update_time",
			Help:      "Total time to pull all data from Dishy",
			Buckets:   prometheus.ExponentialBuckets(1, 3, 8),
		}),
		Updates: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "updates",
			Help:      "Number of Dishy updates",
		}),
		Failures: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "failures",
			Help:      "Number of Dishy request failures",
		}),
		Requests: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "requests",
			Help:      "Number of Dishy requests",
		}),
		Failing: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "exporter",
			Name:      "failing",
			Help:      "Boolean indicator if requests to Dishy are failing",
		}),

		// Dishy info metrics
		Bootcount: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "bootcount",
			Help:      "Dishy Boot Count",
		}, []string{"id", "hardware_version", "software_version", "manufactured_version", "country_code"}),
		UptimeS: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "uptime_s",
			Help:      "Uptime of Dishy",
		}),

		// Dishy GPS metrics
		GPSValid: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "gps_valid",
			Help:      "Boolean indicator for GPS Valid",
		}),
		GPSSats: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "gps_sats",
			Help:      "Number of available GPS Satellites",
		}),

		// Alert metrics
		AlertStatus: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "alert_status",
			Help:      "Status of Alerts",
		}, []string{"alert"}),
		Obstructed: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "obstructed",
			Help:      "Boolean, dishy is obstructed",
		}),
		FractionObstructed: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "fraction_obstructed",
			Help:      "Percent Obstructed",
		}),
		AvgObstructedDurS: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "avg_prolonged_obstruction_duration_s",
			Help:      "Average duration of obstruction",
		}),
		Outage: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "outage",
			Help:      "Boolean, current Dishy outage status",
		}),
		Alerts: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "alerts",
			Help:      "Number of current alerts",
		}),

		// Throughput/latency metrics
		PopPingDropRate: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "pop_ping_drop_rate",
			Help:      "Current pop ping drop rate",
		}),
		PopPingLatencyMs: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "pop_ping_latency_ms",
			Help:      "Current pop ping latency",
		}),
		DLThroughputBps: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "downlink_throughput_bps",
			Help:      "Current downlink throughput in bits per second",
		}),
		ULThroughputBps: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "uplink_throughput_bps",
			Help:      "Current uplink throughput in bits per second",
		}),

		// Orientation metrics
		AzimuthDeg: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "boresight_azimuth_deg",
			Help:      "Boresight azimuth in degrees",
		}),
		ElevationDeg: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "boresight_elevation_deg",
			Help:      "Boresight elevation in degrees",
		}),
		EthSpeedMbps: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "eth_speed_mbps",
			Help:      "Ethernet speed in mbps",
		}),

		// Outage metrics
		OutageHistogram: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "outage_times",
			Help:      "Histogram of outage times",
			Buckets:   prometheus.ExponentialBucketsRange(0.25, 3600, 15),
		}),
		Outages: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "outages",
			Help:      "Number of recent outages",
		}, []string{"cause"}),
		AvgOutageDuration: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "outage_duration_sec_avg",
			Help:      "Avg Outage Duration",
		}, []string{"cause"}),
		SumOutageDuration: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "outage_duration_sec_sum",
			Help:      "Total Outage Duration",
		}, []string{"cause"}),

		// NEW: Additional status metrics
		StowRequested: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "stow_requested",
			Help:      "Boolean, 1 if stow requested",
		}),
		SNRAboveNoiseFloor: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "snr_above_noise_floor",
			Help:      "Boolean, 1 if SNR is healthy above noise floor",
		}),
		SNRPersistentlyLow: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "snr_persistently_low",
			Help:      "Boolean, 1 if SNR is persistently low",
		}),
		SoftwareUpdateState: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "software_update_state",
			Help:      "Software update state (labeled by state name)",
		}, []string{"state"}),
		SoftwareUpdateProgress: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "software_update_progress",
			Help:      "Software update progress (0.0 to 1.0)",
		}),
		TiltAngleDeg: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "tilt_angle_deg",
			Help:      "Dish tilt angle in degrees",
		}),
		AttitudeUncertaintyDeg: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "attitude_uncertainty_deg",
			Help:      "Attitude/alignment uncertainty in degrees",
		}),
		SecondsToFirstNonempty: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "seconds_to_first_nonempty_slot",
			Help:      "Time to next non-empty slot",
		}),
		IsCellDisabled: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "is_cell_disabled",
			Help:      "Boolean, 1 if cell is disabled",
		}),
		MobilityClass: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "mobility_class",
			Help:      "Mobility class (labeled by class name)",
		}, []string{"class"}),

		// NEW: Additional obstruction stats
		TimeObstructedS: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "time_obstructed_s",
			Help:      "Total seconds obstructed",
		}),
		ObstructionIntervalS: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "obstruction_interval_s",
			Help:      "Average interval between obstructions in seconds",
		}),
		ObstructionValidS: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "starlink",
			Subsystem: "dishy",
			Name:      "obstruction_valid_s",
			Help:      "Seconds of valid obstruction data",
		}),
	}

	return m
}
