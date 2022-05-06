package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	starlink "rdmcguire/starlink-status/device"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const timeout = 5

// Setup
var (
	host     string = "192.168.100.1:9000" // Default for Dishy
	promAddr string = "0.0.0.0:9982"       // Listen address for Prometheus
	interval string = "30s"                // Update seconds
	logLevel string = "info"               // Logging Level
)

//Shared Variables
var (
	client      starlink.DeviceClient                         // GRPC Connection to Dishy
	dishy       *starlink.Request     = new(starlink.Request) // Dishy GRPC Request Provider
	log         *logrus.Logger        = logrus.New()          // Logrus logger
	dishyLabels prometheus.Labels
	wg          sync.WaitGroup
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
	log.Debug("Updating Info Metrics")
	updateInfoMetrics()
	log.Debug("Updating Status Metrics")
	updateStatusMetrics()
}

// Requests GetDeviceInfo and updated relevant metrics
func updateInfoMetrics() {
	// Fetch DeviceInfo
	dishy.Request = &starlink.Request_GetDeviceInfo{}
	info, err := getRequest()
	if err != nil {
		log.WithField("Error", err).Warn("Unable to GetDeviceInfo from dishy")
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
		log.WithField("Error", err).Warn("Unable to GetDeviceInfo from dishy")
		return
	}
	dishStatus := status.GetDishGetStatus()

	// GPS Statistics
	var GPSValid int8
	if dishStatus.GetGpsStats().GpsValid {
		GPSValid = 1
	}
	promDishyGPSValid.Set(float64(GPSValid))
	promDishyGPSSats.Set(float64(dishStatus.GetGpsStats().GetGpsSats()))

	// Device Booleans
	var stuck int8
	if dishStatus.Alerts.GetMotorsStuck() {
		stuck = 1
	}
	promDishyMotorsStuck.Set(float64(stuck))

	var hot int8
	if dishStatus.Alerts.GetThermalThrottle() {
		hot = 1
	}
	promDishyThermalThrottle.Set(float64(hot))

	var reallyHot int8
	if dishStatus.Alerts.GetThermalShutdown() {
		reallyHot = 1
	}
	promDishyThermalShutdown.Set(float64(reallyHot))

	var notVert int8
	if dishStatus.Alerts.GetMastNotNearVertical() {
		notVert = 1
	}
	promDishyMastNotNearVertical.Set(float64(notVert))
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
		"id":               info.GetGetDeviceInfo().DeviceInfo.Id,
		"country_code":     info.GetGetDeviceInfo().DeviceInfo.CountryCode,
		"hardware_version": info.GetGetDeviceInfo().DeviceInfo.HardwareVersion,
		"software_version": info.GetGetDeviceInfo().DeviceInfo.SoftwareVersion,
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
	log.Printf("GPS: %+v", status.GetDishGetStatus().GetGpsStats())
	log.Printf("DishObstructed: %+v", status.GetDishGetStatus().GetObstructionStats().GetCurrentlyObstructed())
	log.Printf("DeviceAlerts: %v", status.GetDishGetStatus().GetAlerts())
	log.Printf("MotorsStuck: %v", status.GetDishGetStatus().Alerts.GetMotorsStuck())
	log.Printf("ThermalThrottle: %v", status.GetDishGetStatus().Alerts.GetThermalThrottle())
	log.Printf("ThermalShutdown: %v", status.GetDishGetStatus().Alerts.GetThermalShutdown())
	log.Printf("PopPingDropRate: %+v", status.GetDishGetStatus().PopPingDropRate)
	log.Printf("CurrentElevation: %+v", status.GetDishGetStatus().GetBoresightElevationDeg())
	log.Printf("CurrentAzimuth: %+v", status.GetDishGetStatus().GetBoresightAzimuthDeg())
	log.Printf("Outage: %+v", status.GetDishGetStatus().GetOutage())

	// History
	dishy.Request = &starlink.Request_GetHistory{}
	history, _ := getRequest()
	log.Printf("PopPingDropRateLast20: %+v", history.GetDishGetHistory().PopPingDropRate[len(history.GetDishGetHistory().PopPingDropRate)-20:])
	outages := history.GetDishGetHistory().GetOutages()
	log.Print("Last 5 Outages:")
	for i := 1; i < 6; i++ {
		outage := outages[len(outages)-i]
		outageTime := time.Unix(0, outage.StartTimestampNs)
		log.Printf("\tOutage @ %s [%ds]\tswitched:%v cause:%s", outageTime.Format("15:04:05 MST"), outage.GetDurationNs()/100000000, outage.GetDidSwitch(), outage.GetCause())
	}

	// Config
	dishy.Request = &starlink.Request_DishGetConfig{}
	conf, _ := getRequest()
	log.Printf("Config %+v", conf)
}

// Generic request getter
func getRequest() (*starlink.Response, error) {
	// Prepare request context
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	resp, err := client.Handle(ctx, dishy) // Make request
	//r.Reset()                           // Reset request

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
