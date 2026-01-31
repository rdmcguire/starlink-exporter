package dishy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitea.libretechconsulting.com/50W/starlink-exporter/pkg/device"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

// debugRequests is the list of requests to make in debug mode
var debugRequests = []*device.Request{
	{Request: &device.Request_GetDeviceInfo{}},
	{Request: &device.Request_GetDiagnostics{}},
	{Request: &device.Request_GetHistory{}},
	{Request: &device.Request_GetLocation{}},
	{Request: &device.Request_GetStatus{}},
	{Request: &device.Request_DishGetConfig{}},
	{Request: &device.Request_DishGetObstructionMap{}},
}

// DebugDishy dumps all available Dishy API responses as JSON
func DebugDishy(ctx context.Context, d device.DeviceClient, logger *logrus.Logger) {
	var wg sync.WaitGroup
	var printLock sync.Mutex

	for _, r := range debugRequests {
		wg.Add(1)
		go func(req *device.Request) {
			defer wg.Done()
			dCtx, dCncl := context.WithTimeout(ctx, 5*time.Second)
			defer dCncl()

			logger.Infof("calling %s", req.String())
			getAndPrint(dCtx, d, req, logger, &printLock)
		}(r)
	}
	wg.Wait()
}

func getAndPrint(ctx context.Context, d device.DeviceClient, req *device.Request, logger *logrus.Logger, printLock *sync.Mutex) {
	resp, e := d.Handle(ctx, req)
	if e != nil {
		logger.Error(e.Error(), " ::request: ", req.String())
		return
	}

	printLock.Lock()
	fmt.Printf("\n+ %s\n%s\n",
		req.String(),
		protojson.Format(resp))
	printLock.Unlock()
}

// DumpData logs detailed debug information about Dishy state
func DumpData(ctx context.Context, c *Client) {
	logger := c.logger

	// Device Info
	info, err := c.GetDeviceInfo(ctx)
	if err == nil {
		logger.Printf("Info: %+v", info.GetGetDeviceInfo().DeviceInfo)
	}

	// Status
	status, err := c.GetStatus(ctx)
	if err == nil {
		dishStatus := status.GetDishGetStatus()
		logger.Debugf("GPS: %+v", dishStatus.GetGpsStats())
		logger.Debugf("DishObstructed: %+v", dishStatus.GetObstructionStats().GetCurrentlyObstructed())
		logger.Debugf("DeviceAlerts: %v", dishStatus.GetAlerts())
		logger.Debugf("MotorsStuck: %v", dishStatus.Alerts.GetMotorsStuck())
		logger.Debugf("ThermalThrottle: %v", dishStatus.Alerts.GetThermalThrottle())
		logger.Debugf("ThermalShutdown: %v", dishStatus.Alerts.GetThermalShutdown())
		logger.Debugf("PopPingDropRate: %+v", dishStatus.PopPingDropRate)
		logger.Debugf("CurrentElevation: %+v", dishStatus.GetBoresightElevationDeg())
		logger.Debugf("CurrentAzimuth: %+v", dishStatus.GetBoresightAzimuthDeg())
		logger.Debugf("Outage: %+v", dishStatus.GetOutage())
	}

	// History
	history, err := c.GetHistory(ctx)
	if err == nil {
		histData := history.GetDishGetHistory()
		if len(histData.PopPingDropRate) >= 20 {
			logger.Debugf("PopPingDropRateLast20: %+v", histData.PopPingDropRate[len(histData.PopPingDropRate)-20:])
		}
		outages := histData.GetOutages()
		logger.Debugf("Current %+v", histData.GetCurrent())
		logger.Debug("Outages:")
		for i := range outages {
			outage := outages[i]
			outageTime := time.Unix(0, outage.StartTimestampNs)
			logger.Debugf("\tOutage @ %s [%ds]\tswitched:%v cause:%s",
				outageTime.Format("15:04:05 MST"),
				outage.GetDurationNs()/100000000,
				outage.GetDidSwitch(),
				outage.GetCause())
		}
	}
}
