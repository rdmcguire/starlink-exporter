package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "rdmcguire/starlink-status/device"

	"google.golang.org/grpc"
)

const timeout = 5

var (
	dishy    pb.DeviceClient
	host     string      = "192.168.100.1:9000" // Default for Dishy
	interval int         = 5                    // Update seconds
	r        *pb.Request = new(pb.Request)
)

func init() {
	flag.StringVar(&host, "host", host, "IP and port of Dishy GRPC endpoint")
	flag.IntVar(&interval, "interval", interval, "Update interval in seconds")
	flag.Parse()
}

func main() {
	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "192.168.100.1:9200", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to dishy: %+v", err)
	}
	dishy = pb.NewDeviceClient(conn)
	defer conn.Close()

	// Get Info
	r.Request = &pb.Request_GetDeviceInfo{}
	info, _ := getRequest(r)
	log.Printf("Info: %+v", info.GetGetDeviceInfo().DeviceInfo)

	// Get Status
	r.Request = &pb.Request_GetStatus{}
	status, _ := getRequest(r)
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
	r.Request = &pb.Request_GetHistory{}
	history, _ := getRequest(r)
	log.Printf("PopPingDropRateLast20: %+v", history.GetDishGetHistory().PopPingDropRate[len(history.GetDishGetHistory().PopPingDropRate)-20:])
	outages := history.GetDishGetHistory().GetOutages()
	log.Print("Last 5 Outages:")
	for i := 1; i < 6; i++ {
		outage := outages[len(outages)-i]
		outageTime := time.Unix(0, outage.StartTimestampNs)
		log.Printf("\tOutage @ %s [%ds]\tswitched:%v cause:%s", outageTime.Format("15:04:05 MST"), outage.GetDurationNs()/100000000, outage.GetDidSwitch(), outage.GetCause())
	}

	// Test
	r.Request = &pb.Request_DishGetConfig{}
	conf, _ := getRequest(r)
	log.Printf("Config %+v", conf)

}

// Generic request getter
func getRequest(req *pb.Request) (*pb.Response, error) {
	// Prepare request context
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	resp, err := dishy.Handle(ctx, req) // Make request
	//r.Reset()                           // Reset request

	return resp, err
}
