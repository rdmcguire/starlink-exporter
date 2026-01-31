package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitea.libretechconsulting.com/50W/starlink-exporter/internal/config"
	"gitea.libretechconsulting.com/50W/starlink-exporter/internal/dishy"
	"gitea.libretechconsulting.com/50W/starlink-exporter/internal/exporter"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse configuration
	cfg, err := config.Parse()
	if err != nil {
		logrus.Fatalf("Failed to parse configuration: %v", err)
	}
	log := cfg.Logger

	// Setup signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Connect to Dishy
	client, err := dishy.NewClient(cfg.Host, log)
	if err != nil {
		log.Fatalf("Failed to connect to dishy: %v", err)
	}
	defer client.Close()

	log.Info("GRPC Connected to Dishy")

	// Handle debug mode
	if cfg.DebugInfo {
		dishy.DebugDishy(ctx, client.DeviceClient(), log)
		return
	}

	// Initialize metrics
	metrics := exporter.NewMetrics()

	// Start Prometheus server
	go func() {
		if err := exporter.StartServer(cfg.PromAddr, metrics, log); err != nil {
			log.WithField("Error", err).Fatal("Failed to start Prometheus")
		}
	}()

	// Initialize collector
	collector := exporter.NewCollector(client, metrics, log)

	// Initialize labels from device info
	if err := collector.InitLabels(ctx); err != nil {
		log.Warnf("Failed to initialize labels: %v", err)
	}

	// Dump debug info if debug level enabled
	if log.IsLevelEnabled(logrus.DebugLevel) {
		dishy.DumpData(ctx, client)
	}

	// Create ticker for update loop
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Info("Serving metrics, entering update loop")

	// Initial update (don't wait for first tick)
	collector.UpdateMetrics(ctx)

	// Update loop
	for {
		select {
		case <-ctx.Done():
			log.Warn("Asked to die, waiting on goroutines...", "reason", ctx.Err())
			collector.Wait()
			os.Exit(0)
		case <-ticker.C:
			collector.UpdateMetrics(ctx)
		}
	}
}
