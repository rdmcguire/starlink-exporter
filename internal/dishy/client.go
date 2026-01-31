package dishy

import (
	"context"
	"time"

	"gitea.libretechconsulting.com/50W/starlink-exporter/pkg/device"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RequestTimeout is the default timeout for gRPC requests
const RequestTimeout = 5 * time.Second

// Client wraps the gRPC device client with helper methods
type Client struct {
	conn   *grpc.ClientConn
	device device.DeviceClient
	logger *logrus.Logger

	// Metrics callbacks for request instrumentation
	OnRequestComplete func(durationMs float64)
	OnRequestFailure  func()
	OnRequestSuccess  func()
}

// NewClient creates a new Dishy gRPC client
func NewClient(host string, logger *logrus.Logger) (*Client, error) {
	conn, err := grpc.NewClient(host,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		device: device.NewDeviceClient(conn),
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// DeviceClient returns the underlying device client for debug operations
func (c *Client) DeviceClient() device.DeviceClient {
	return c.device
}

// GetDeviceInfo fetches device information
func (c *Client) GetDeviceInfo(ctx context.Context) (*device.Response, error) {
	return c.request(ctx, &device.Request{
		Request: &device.Request_GetDeviceInfo{},
	})
}

// GetStatus fetches current dish status
func (c *Client) GetStatus(ctx context.Context) (*device.Response, error) {
	return c.request(ctx, &device.Request{
		Request: &device.Request_GetStatus{},
	})
}

// GetHistory fetches dish history including outages
func (c *Client) GetHistory(ctx context.Context) (*device.Response, error) {
	return c.request(ctx, &device.Request{
		Request: &device.Request_GetHistory{},
	})
}

// request is the generic request helper
func (c *Client) request(ctx context.Context, req *device.Request) (*device.Response, error) {
	reqCtx, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()

	t1 := time.Now()
	resp, err := c.device.Handle(reqCtx, req)
	durationMs := float64(time.Since(t1).Milliseconds())

	if c.OnRequestComplete != nil {
		c.OnRequestComplete(durationMs)
	}

	if err != nil {
		if c.OnRequestFailure != nil {
			c.OnRequestFailure()
		}
		c.logger.WithFields(logrus.Fields{
			"Request": req.String(),
			"Error":   err,
		}).Error("Unable to request data from Dishy")
		return nil, err
	}

	if c.OnRequestSuccess != nil {
		c.OnRequestSuccess()
	}

	return resp, nil
}
