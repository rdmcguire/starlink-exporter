package config

import (
	"flag"
	"time"

	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	Host      string
	PromAddr  string
	Interval  time.Duration
	LogLevel  string
	DebugInfo bool
	Logger    *logrus.Logger
}

// Default values
const (
	DefaultHost     = "192.168.100.1:9200"
	DefaultPromAddr = "0.0.0.0:9982"
	DefaultInterval = "30s"
	DefaultLogLevel = "info"
)

// Parse parses command line flags and returns a Config
func Parse() (*Config, error) {
	host := flag.String("host", DefaultHost, "IP and port of Dishy GRPC endpoint")
	interval := flag.String("interval", DefaultInterval, "Update interval (go time.Duration e.g. 1m30s)")
	promAddr := flag.String("promAddr", DefaultPromAddr, "Listen address and port for Prometheus /metrics")
	logLevel := flag.String("logLevel", DefaultLogLevel, "Logging level (error, warn, info, debug, trace)")
	debugInfo := flag.Bool("debug", false, "Dump dishy info")
	flag.Parse()

	// Parse duration
	duration, err := time.ParseDuration(*interval)
	if err != nil {
		return nil, err
	}

	// Create logger
	logger := logrus.New()
	setLogLevel(logger, *logLevel)

	return &Config{
		Host:      *host,
		PromAddr:  *promAddr,
		Interval:  duration,
		LogLevel:  *logLevel,
		DebugInfo: *debugInfo,
		Logger:    logger,
	}, nil
}

// setLogLevel sets the log level on the logger
func setLogLevel(logger *logrus.Logger, level string) {
	switch level {
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}
