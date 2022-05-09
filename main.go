package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	starlink "rdmcguire/starlink-exporter/device"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// GRPC Timeout
const timeout = 5

// List of alerts by field name
var alerts = []string{
	"MotorsStuck",
	"ThermalThrottle",
	"ThermalShutdown",
	"MastNotNearVertical",
	"UnexpectedLocation",
	"SlowEthernetSpeeds",
	"Roaming",
}

// Setup
var (
	host     string = "192.168.100.1:9000" // Default for Dishy
	promAddr string = "0.0.0.0:9982"       // Listen address for Prometheus
	interval string = "30s"                // Update seconds
	logLevel string = "info"               // Logging Level
)

//Shared Variables
var (
	client       starlink.DeviceClient                         // GRPC Connection to Dishy
	dishy        *starlink.Request     = new(starlink.Request) // Dishy GRPC Request Provider
	log          *logrus.Logger        = logrus.New()          // Logrus logger
	dishyLabels  prometheus.Labels
	wg           sync.WaitGroup
	latestOutage int64
)

func init() {
	// Handle flags
	flag.StringVar(&host, "host", host, "IP and port of Dishy GRPC endpoint")
	flag.StringVar(&interval, "interval", interval, "Update interval (go time.Duration e.g. 1m30s)")
	flag.StringVar(&promAddr, "promAddr", promAddr, "Listen address and port for Prometheus /metrics")
	flag.StringVar(&logLevel, "logLevel", logLevel, "Logging level (error, warn, info, debug, trace)")
	flag.Parse()

	// Set logging
	setLogLevel()
}

func UpdateMetrics() {
	wg.Add(1)
	defer wg.Done()

	t1 := time.Now()
	log.Debug("Updating Metrics")

	log.Trace("Updating Info Metrics")
	updateInfoMetrics()
	log.Trace("Updating Status Metrics")
	updateStatusMetrics()
	log.Trace("Updating History Metrics")
	updateHistoryMetrics()

	promDishyUpdates.Inc()
	promDishyUpdateTime.Observe(float64(time.Now().Sub(t1).Milliseconds()))
}

// Requests GetDeviceInfo and updated relevant metrics
func updateInfoMetrics() {
	// Fetch DeviceInfo
	dishy.Request = &starlink.Request_GetDeviceInfo{}
	info, err := getRequest()
	if err != nil {
		return
	}

	// Boot Count
	promDishyBootcount.With(dishyLabels).
		Set(float64(info.GetGetDeviceInfo().DeviceInfo.Bootcount))
}

func updateStatusMetrics() {
	// Fetch Dishy Status
	dishy.Request = &starlink.Request_GetStatus{}
	status, err := getRequest()
	if err != nil {
		return
	}
	dishStatus := status.GetDishGetStatus()

	// GPS Statistics
	var GPSValid float64
	if dishStatus.GetGpsStats().GpsValid {
		GPSValid = 1
	}
	promDishyGPSValid.Set(GPSValid)
	promDishyGPSSats.Set(float64(dishStatus.GetGpsStats().GetGpsSats()))

	// Currently In Outage
	var inOutage float64
	if dishStatus.Outage != nil {
		inOutage = 1
	}
	promDishyOutage.Set(inOutage)

	// Current Obstructed State
	var obstructed float64
	if dishStatus.ObstructionStats.CurrentlyObstructed {
		obstructed = 1
	}
	promDishyObstructed.Set(obstructed)

	// Device Alert Booleans
	for _, name := range alerts {
		promDishyAlertStatus.WithLabelValues(name).
			Set(isAlerting(dishStatus.Alerts, name))
	}

	// Status Metrics
	promDishyUptimeS.Set(float64(dishStatus.GetDeviceState().GetUptimeS()))
	promDishyAlerts.Set(countAlerts(dishStatus.Alerts))
	promDishyFractionObstructed.Set(float64(dishStatus.ObstructionStats.GetFractionObstructed()))
	promDishyAvgObstructedDurationS.Set(float64(dishStatus.GetObstructionStats().GetAvgProlongedObstructionDurationS()))
	promDishyPopPingDropRate.Set(float64(dishStatus.GetPopPingDropRate()))
	promDishyPopPingLatencyMs.Set(float64(dishStatus.GetPopPingLatencyMs()))
	promDishyDLTputBps.Set(float64(dishStatus.GetDownlinkThroughputBps()))
	promDishyULTputBps.Set(float64(dishStatus.GetUplinkThroughputBps()))
	promDishyAzimuthDeg.Set(float64(dishStatus.GetBoresightAzimuthDeg()))
	promDishyElevationDeg.Set(float64(dishStatus.GetBoresightElevationDeg()))
	promDishyEthSpeedMbps.Set(float64(dishStatus.GetEthSpeedMbps()))
}

// Update History Metricis
func updateHistoryMetrics() {
	// Fetch History Metrics
	dishy.Request = &starlink.Request_GetHistory{}
	history, err := getRequest()
	if err != nil {
		return
	}

	// Outage History
	outages := history.GetDishGetHistory().GetOutages()

	// Outage Histogram
	// Steps backwards, observing any newly seen outages
	for i := len(outages) - 1; i >= 0; i-- {
		if outages[i].GetStartTimestampNs() > latestOutage {
			promDishyOutageHistogram.Observe(float64(outages[i].GetDurationNs() / 1e9))
		} else {
			break
		}
	}
	// Advance our latest timestamp
	latestOutage = outages[len(outages)-1].GetStartTimestampNs()

	// Calculate Count/Sum/Avg Outage Durations by Cause
	durationSums := make(map[string]float64)
	durationCounts := make(map[string]float64)
	for _, outage := range outages {
		durationSums[outage.GetCause().String()] += float64(outage.GetDurationNs() / 1e9)
		durationCounts[outage.GetCause().String()]++
	}
	for cause := range durationSums {
		promDishyAvgOutageDuration.WithLabelValues(cause).
			Set(durationSums[cause] / durationCounts[cause])
		promDishySumOutageDuration.WithLabelValues(cause).
			Set(durationSums[cause])
		promDishyOutages.WithLabelValues(cause).
			Set(durationCounts[cause])
	}
}

// Returns the status of an alert as float64
func isAlerting(a *starlink.DishAlerts, alert string) float64 {
	var firing float64
	r := reflect.ValueOf(a)
	if reflect.Indirect(r).FieldByName(alert).Bool() {
		firing = 1
	}
	return firing
}

// Counts the number of alerts currently activated
func countAlerts(a *starlink.DishAlerts) float64 {
	var firing float64
	r := reflect.ValueOf(a)
	for _, alert := range alerts {
		if reflect.Indirect(r).FieldByName(alert).Bool() {
			firing++
		}
	}
	return firing
}

func main() {
	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "192.168.100.1:9200", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to dishy: %+v", err)
	}
	client = starlink.NewDeviceClient(conn)
	defer conn.Close()

	log.Info("GRPC Connected to Dishy")

	// Prepare Prometheus
	go promInit()

	// Get Info
	dishy.Request = &starlink.Request_GetDeviceInfo{}
	info, _ := getRequest()

	// Prepare Dishy info labels
	dishyLabels = prometheus.Labels{
		"id":                   info.GetGetDeviceInfo().DeviceInfo.Id,
		"country_code":         info.GetGetDeviceInfo().DeviceInfo.CountryCode,
		"hardware_version":     info.GetGetDeviceInfo().DeviceInfo.HardwareVersion,
		"software_version":     info.GetGetDeviceInfo().DeviceInfo.SoftwareVersion,
		"manufactured_version": info.GetGetDeviceInfo().DeviceInfo.ManufacturedVersion,
	}

	// Dump some stats if debug
	if log.IsLevelEnabled(logrus.DebugLevel) {
		dumpData()
	}

	// Handle death
	die := make(chan os.Signal)
	signal.Notify(die, syscall.SIGINT, syscall.SIGTERM)

	// Parse duration and create a ticker
	duration, err := time.ParseDuration(interval)
	if err != nil {
		log.WithFields(logrus.Fields{"Interval": interval, "Error": err}).
			Fatal("Failed to parse interval")
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	log.Info("Serving metrics, entering update loop")

	UpdateMetrics() // Don't wait for the first Tick

	// Update forever
	for {
		select {
		case <-die:
			log.Warn("Asked to die, waiting on goroutines...")
			wg.Wait()
			os.Exit(0)
		case <-ticker.C:
			UpdateMetrics()
		}
	}

}

// Dump some info if level is high enough
func dumpData() {
	// Device Info
	dishy.Request = &starlink.Request_GetDeviceInfo{}
	info, _ := getRequest()
	log.Printf("Info: %+v", info.GetGetDeviceInfo().DeviceInfo)

	// Status
	dishy.Request = &starlink.Request_GetStatus{}
	status, _ := getRequest()
	log.Debugf("GPS: %+v", status.GetDishGetStatus().GetGpsStats())
	log.Debugf("DishObstructed: %+v", status.GetDishGetStatus().GetObstructionStats().GetCurrentlyObstructed())
	log.Debugf("DeviceAlerts: %v", status.GetDishGetStatus().GetAlerts())
	log.Debugf("MotorsStuck: %v", status.GetDishGetStatus().Alerts.GetMotorsStuck())
	log.Debugf("ThermalThrottle: %v", status.GetDishGetStatus().Alerts.GetThermalThrottle())
	log.Debugf("ThermalShutdown: %v", status.GetDishGetStatus().Alerts.GetThermalShutdown())
	log.Debugf("PopPingDropRate: %+v", status.GetDishGetStatus().PopPingDropRate)
	log.Debugf("CurrentElevation: %+v", status.GetDishGetStatus().GetBoresightElevationDeg())
	log.Debugf("CurrentAzimuth: %+v", status.GetDishGetStatus().GetBoresightAzimuthDeg())
	log.Debugf("Outage: %+v", status.GetDishGetStatus().GetOutage())

	// History
	dishy.Request = &starlink.Request_GetHistory{}
	history, _ := getRequest()
	log.Debugf("PopPingDropRateLast20: %+v", history.GetDishGetHistory().PopPingDropRate[len(history.GetDishGetHistory().PopPingDropRate)-20:])
	outages := history.GetDishGetHistory().GetOutages()
	log.Debugf("Current %+v", history.GetDishGetHistory().GetCurrent())
	log.Debug("Outages:")
	for i := 0; i < len(outages); i++ {
		outage := outages[i]
		outageTime := time.Unix(0, outage.StartTimestampNs)
		log.Debugf("\tOutage @ %s [%ds]\tswitched:%v cause:%s", outageTime.Format("15:04:05 MST"), outage.GetDurationNs()/100000000, outage.GetDidSwitch(), outage.GetCause())
	}

	// Config
	dishy.Request = &starlink.Request_DishGetConfig{}
	conf, _ := getRequest()
	log.Debugf("Config %+v", conf)
}

// Generic request getter
func getRequest() (*starlink.Response, error) {
	// Prepare request context
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	t1 := time.Now()
	resp, err := client.Handle(ctx, dishy) // Make request

	promDishyGRPCTime.Observe(float64(time.Now().Sub(t1).Milliseconds()))

	if err != nil {
		promDishyFailing.Set(1)
		promDishyFailures.Inc()
		log.WithFields(logrus.Fields{
			"Request": dishy.Request,
			"Error":   err,
		}).Error("Unable to request data from Dishy")
	} else {
		promDishyFailing.Set(0)
	}

	promDishyRequests.Inc()

	return resp, err
}

// Check for log level in config, use default if not found
func setLogLevel() {
	switch logLevel {
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}
