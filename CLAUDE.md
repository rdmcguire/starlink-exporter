# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Starlink-exporter is a Prometheus exporter for SpaceX's Starlink Dishy satellite terminal. It collects metrics via gRPC from the Dishy device and exposes them on a Prometheus HTTP endpoint.

## Build Commands

```bash
# Build the binary
go build -o starlink-exporter ./cmd/starlink-exporter

# Run the exporter (requires network access to Dishy at 192.168.100.1)
./starlink-exporter

# Run with flags
./starlink-exporter --host 192.168.100.1:9200 --promAddr 0.0.0.0:9982 --interval 30s --logLevel info

# Debug mode - dumps all Dishy API responses as JSON
./starlink-exporter --debug

# Regenerate protobuf code from dish.protoset (requires protoc installed)
./contrib/regen.sh

# Install protoc compiler
./contrib/get-protoc.sh
```

## Architecture

### Directory Structure
```
cmd/starlink-exporter/
  main.go              - Entry point, wires up dependencies

internal/
  config/
    config.go          - Flag parsing, Config struct

  dishy/
    client.go          - gRPC client wrapper, connection management
    requests.go        - Request helpers, debug dump functions

  exporter/
    metrics.go         - Prometheus metric definitions (gauges, counters, histograms)
    collectors.go      - Metric update logic (UpdateMetrics, updateStatusMetrics, etc.)
    server.go          - HTTP server for /metrics endpoint
    alerts.go          - Alert introspection helpers (IsAlerting, CountAlerts)
```

### Generated Code (pkg/)
- **pkg/device/**: gRPC client stubs and proto messages for Dishy device API
- **pkg/spacex/**: Supporting proto messages (satellites, telemetron dependencies)

Generated from `proto/dish.protoset` via `contrib/regen.sh`. The protoset is a reflected descriptor set from the actual Dishy device.

### Metrics Flow
1. gRPC client connects to Dishy at `192.168.100.1:9200` (insecure)
2. Update loop calls three endpoints: GetDeviceInfo, GetStatus, GetHistory
3. Metrics exposed on HTTP endpoint (default `:9982/metrics`)

### Metric Namespaces
- `starlink_exporter_*`: Internal exporter health (grpc_time, update_time, failures)
- `starlink_dishy_*`: Dishy operational metrics (gps, alerts, throughput, latency, outages)

## Key Technical Details

- Uses modern `grpc.NewClient()` API (not deprecated DialContext)
- Alert types are introspected via reflection from the proto message
- Outage causes are collected from GetHistory with duration statistics
- Graceful shutdown via SIGINT/SIGTERM signal handling

## Docker

Multi-stage build with Alpine runtime. No special permissions or volumes needed:
```bash
docker build -t starlink-exporter .
docker run -p 9982:9982 starlink-exporter
```

## CI/CD

Gitea Actions workflow (`.gitea/workflows/build.yaml`) builds and pushes Docker images on push events. Tags images with branch name and `latest`.
