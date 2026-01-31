package exporter

import (
	"context"
	"sync"
	"time"

	"gitea.libretechconsulting.com/50W/starlink-exporter/internal/dishy"
	"gitea.libretechconsulting.com/50W/starlink-exporter/pkg/device"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Collector handles metric collection from Dishy
type Collector struct {
	client       *dishy.Client
	metrics      *Metrics
	logger       *logrus.Logger
	dishyLabels  prometheus.Labels
	latestOutage int64
	wg           sync.WaitGroup
}

// NewCollector creates a new metric collector
func NewCollector(client *dishy.Client, metrics *Metrics, logger *logrus.Logger) *Collector {
	c := &Collector{
		client:  client,
		metrics: metrics,
		logger:  logger,
	}

	// Wire up client callbacks for request instrumentation
	client.OnRequestComplete = func(durationMs float64) {
		metrics.GRPCTime.Observe(durationMs)
		metrics.Requests.Inc()
	}
	client.OnRequestFailure = func() {
		metrics.Failing.Set(1)
		metrics.Failures.Inc()
	}
	client.OnRequestSuccess = func() {
		metrics.Failing.Set(0)
	}

	return c
}

// InitLabels fetches device info and initializes labels
func (c *Collector) InitLabels(ctx context.Context) error {
	info, err := c.client.GetDeviceInfo(ctx)
	if err != nil {
		return err
	}

	c.dishyLabels = prometheus.Labels{
		"id":                   info.GetGetDeviceInfo().DeviceInfo.Id,
		"country_code":         info.GetGetDeviceInfo().DeviceInfo.CountryCode,
		"hardware_version":     info.GetGetDeviceInfo().DeviceInfo.HardwareVersion,
		"software_version":     info.GetGetDeviceInfo().DeviceInfo.SoftwareVersion,
		"manufactured_version": info.GetGetDeviceInfo().DeviceInfo.ManufacturedVersion,
	}

	return nil
}

// Wait waits for any in-progress metric updates to complete
func (c *Collector) Wait() {
	c.wg.Wait()
}

// UpdateMetrics fetches all metrics from Dishy
func (c *Collector) UpdateMetrics(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	t1 := time.Now()
	c.logger.Debug("Updating Metrics")

	c.logger.Trace("Updating Info Metrics")
	c.updateInfoMetrics(ctx)

	c.logger.Trace("Updating Status Metrics")
	c.updateStatusMetrics(ctx)

	c.logger.Trace("Updating History Metrics")
	c.updateHistoryMetrics(ctx)

	c.metrics.Updates.Inc()
	c.metrics.UpdateTime.Observe(float64(time.Since(t1).Milliseconds()))
}

// updateInfoMetrics fetches GetDeviceInfo and updates relevant metrics
func (c *Collector) updateInfoMetrics(ctx context.Context) {
	info, err := c.client.GetDeviceInfo(ctx)
	if err != nil {
		return
	}

	c.metrics.Bootcount.With(c.dishyLabels).
		Set(float64(info.GetGetDeviceInfo().DeviceInfo.Bootcount))
}

// updateStatusMetrics fetches GetStatus and updates relevant metrics
func (c *Collector) updateStatusMetrics(ctx context.Context) {
	status, err := c.client.GetStatus(ctx)
	if err != nil {
		return
	}
	dishStatus := status.GetDishGetStatus()

	// GPS Statistics
	var gpsValid float64
	if dishStatus.GetGpsStats().GpsValid {
		gpsValid = 1
	}
	c.metrics.GPSValid.Set(gpsValid)
	c.metrics.GPSSats.Set(float64(dishStatus.GetGpsStats().GetGpsSats()))

	// Currently In Outage
	var inOutage float64
	if dishStatus.Outage != nil {
		inOutage = 1
	}
	c.metrics.Outage.Set(inOutage)

	// Current Obstructed State
	var obstructed float64
	if dishStatus.ObstructionStats.CurrentlyObstructed {
		obstructed = 1
	}
	c.metrics.Obstructed.Set(obstructed)

	// Device Alert Booleans
	for _, name := range AlertNames {
		c.metrics.AlertStatus.WithLabelValues(name).
			Set(IsAlerting(dishStatus.Alerts, name))
	}

	// Status Metrics
	c.metrics.UptimeS.Set(float64(dishStatus.GetDeviceState().GetUptimeS()))
	c.metrics.Alerts.Set(CountAlerts(dishStatus.Alerts))
	c.metrics.FractionObstructed.Set(float64(dishStatus.ObstructionStats.GetFractionObstructed()))
	c.metrics.AvgObstructedDurS.Set(float64(dishStatus.GetObstructionStats().GetAvgProlongedObstructionDurationS()))
	c.metrics.PopPingDropRate.Set(float64(dishStatus.GetPopPingDropRate()))
	c.metrics.PopPingLatencyMs.Set(float64(dishStatus.GetPopPingLatencyMs()))
	c.metrics.DLThroughputBps.Set(float64(dishStatus.GetDownlinkThroughputBps()))
	c.metrics.ULThroughputBps.Set(float64(dishStatus.GetUplinkThroughputBps()))
	c.metrics.AzimuthDeg.Set(float64(dishStatus.GetBoresightAzimuthDeg()))
	c.metrics.ElevationDeg.Set(float64(dishStatus.GetBoresightElevationDeg()))
	c.metrics.EthSpeedMbps.Set(float64(dishStatus.GetEthSpeedMbps()))

	// NEW: Additional status metrics
	c.updateNewStatusMetrics(dishStatus)
}

// updateNewStatusMetrics updates the new status metrics
func (c *Collector) updateNewStatusMetrics(dishStatus *device.DishGetStatusResponse) {
	// Stow requested
	var stowRequested float64
	if dishStatus.GetStowRequested() {
		stowRequested = 1
	}
	c.metrics.StowRequested.Set(stowRequested)

	// SNR above noise floor
	var snrAbove float64
	if dishStatus.GetIsSnrAboveNoiseFloor() {
		snrAbove = 1
	}
	c.metrics.SNRAboveNoiseFloor.Set(snrAbove)

	// SNR persistently low
	var snrLow float64
	if dishStatus.GetIsSnrPersistentlyLow() {
		snrLow = 1
	}
	c.metrics.SNRPersistentlyLow.Set(snrLow)

	// Software update state (labeled)
	updateState := dishStatus.GetSoftwareUpdateState()
	c.metrics.SoftwareUpdateState.Reset()
	c.metrics.SoftwareUpdateState.WithLabelValues(updateState.String()).Set(1)

	// Software update progress from stats
	if stats := dishStatus.GetSoftwareUpdateStats(); stats != nil {
		c.metrics.SoftwareUpdateProgress.Set(float64(stats.GetSoftwareUpdateProgress()))
	}

	// Alignment stats
	if alignStats := dishStatus.GetAlignmentStats(); alignStats != nil {
		c.metrics.TiltAngleDeg.Set(float64(alignStats.GetTiltAngleDeg()))
		c.metrics.AttitudeUncertaintyDeg.Set(float64(alignStats.GetAttitudeUncertaintyDeg()))
	}

	// Seconds to first nonempty slot
	c.metrics.SecondsToFirstNonempty.Set(float64(dishStatus.GetSecondsToFirstNonemptySlot()))

	// Cell disabled
	var cellDisabled float64
	if dishStatus.GetIsCellDisabled() {
		cellDisabled = 1
	}
	c.metrics.IsCellDisabled.Set(cellDisabled)

	// Mobility class (labeled)
	mobilityClass := dishStatus.GetMobilityClass()
	c.metrics.MobilityClass.Reset()
	c.metrics.MobilityClass.WithLabelValues(mobilityClass.String()).Set(1)

	// NEW: Additional obstruction stats
	if obStats := dishStatus.GetObstructionStats(); obStats != nil {
		c.metrics.TimeObstructedS.Set(float64(obStats.GetTimeObstructed()))
		c.metrics.ObstructionIntervalS.Set(float64(obStats.GetAvgProlongedObstructionIntervalS()))
		c.metrics.ObstructionValidS.Set(float64(obStats.GetValidS()))
	}
}

// updateHistoryMetrics fetches GetHistory and updates outage metrics
func (c *Collector) updateHistoryMetrics(ctx context.Context) {
	history, err := c.client.GetHistory(ctx)
	if err != nil {
		return
	}

	outages := history.GetDishGetHistory().GetOutages()
	if len(outages) == 0 {
		return
	}

	// Outage Histogram
	// Steps backwards, observing any newly seen outages
	for i := len(outages) - 1; i >= 0; i-- {
		if outages[i].GetStartTimestampNs() > c.latestOutage {
			c.metrics.OutageHistogram.Observe(float64(outages[i].GetDurationNs() / 1e9))
		} else {
			break
		}
	}
	// Advance our latest timestamp
	c.latestOutage = outages[len(outages)-1].GetStartTimestampNs()

	// Calculate Count/Sum/Avg Outage Durations by Cause
	durationSums := make(map[string]float64)
	durationCounts := make(map[string]float64)
	for _, outage := range outages {
		durationSums[outage.GetCause().String()] += float64(outage.GetDurationNs() / 1e9)
		durationCounts[outage.GetCause().String()]++
	}
	for cause := range durationSums {
		c.metrics.AvgOutageDuration.WithLabelValues(cause).
			Set(durationSums[cause] / durationCounts[cause])
		c.metrics.SumOutageDuration.WithLabelValues(cause).
			Set(durationSums[cause])
		c.metrics.Outages.WithLabelValues(cause).
			Set(durationCounts[cause])
	}
}
