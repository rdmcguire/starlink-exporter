#!/bin/bash
set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Project root is one level up from contrib/
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

PROTOSET="${PROJECT_ROOT}/proto/dish.protoset"

MODULE="gitea.libretechconsulting.com/50W/starlink-exporter"

# Proto to Go package mappings
# Main protos we generate
GO_OPT="module=${MODULE}"
GO_OPT="${GO_OPT},Mspacex_api/device/device.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/command.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/common.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/dish.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/dish_config.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/wifi.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/wifi_config.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/wifi_util.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/device/transceiver.proto=${MODULE}/pkg/device"
GO_OPT="${GO_OPT},Mspacex_api/common/status/status.proto=${MODULE}/pkg/status"
# Dependencies (not generated, but needed for import resolution)
GO_OPT="${GO_OPT},Mspacex_api/satellites/network/ut_disablement_codes.proto=${MODULE}/pkg/spacex/satellites/network"
GO_OPT="${GO_OPT},Mspacex_api/telemetron/public/common/time.proto=${MODULE}/pkg/spacex/telemetron/common"
GO_OPT="${GO_OPT},Mspacex_api/telemetron/public/integrations/rate_limit_reason.proto=${MODULE}/pkg/spacex/telemetron/integrations"
GO_OPT="${GO_OPT},Mspacex_api/device/services/unlock/service.proto=${MODULE}/pkg/spacex/device/services/unlock"
GO_OPT="${GO_OPT},Mspacex_api/device/rssi_scan.proto=${MODULE}/pkg/spacex/device/rssi_scan"
GO_OPT="${GO_OPT},Mspacex_api/device/account_shard.proto=${MODULE}/pkg/spacex/device/account_shard"

GRPC_OPT="module=${MODULE}"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/device.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/command.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/common.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/dish.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/dish_config.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/wifi.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/wifi_config.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/wifi_util.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/transceiver.proto=${MODULE}/pkg/device"
GRPC_OPT="${GRPC_OPT},Mspacex_api/common/status/status.proto=${MODULE}/pkg/status"
GRPC_OPT="${GRPC_OPT},Mspacex_api/satellites/network/ut_disablement_codes.proto=${MODULE}/pkg/spacex/satellites/network"
GRPC_OPT="${GRPC_OPT},Mspacex_api/telemetron/public/common/time.proto=${MODULE}/pkg/spacex/telemetron/common"
GRPC_OPT="${GRPC_OPT},Mspacex_api/telemetron/public/integrations/rate_limit_reason.proto=${MODULE}/pkg/spacex/telemetron/integrations"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/services/unlock/service.proto=${MODULE}/pkg/spacex/device/services/unlock"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/rssi_scan.proto=${MODULE}/pkg/spacex/device/rssi_scan"
GRPC_OPT="${GRPC_OPT},Mspacex_api/device/account_shard.proto=${MODULE}/pkg/spacex/device/account_shard"

# Generate all protos at once to project root - module= strips the prefix
protoc --go_out="${PROJECT_ROOT}" --go_opt="${GO_OPT}" \
       --go-grpc_out="${PROJECT_ROOT}" --go-grpc_opt="${GRPC_OPT}" \
       --descriptor_set_in="${PROTOSET}" \
       spacex_api/device/device.proto \
       spacex_api/device/command.proto \
       spacex_api/device/common.proto \
       spacex_api/device/dish.proto \
       spacex_api/device/dish_config.proto \
       spacex_api/device/wifi.proto \
       spacex_api/device/wifi_config.proto \
       spacex_api/device/wifi_util.proto \
       spacex_api/device/transceiver.proto \
       spacex_api/device/rssi_scan.proto \
       spacex_api/device/account_shard.proto \
       spacex_api/device/services/unlock/service.proto \
       spacex_api/satellites/network/ut_disablement_codes.proto \
       spacex_api/telemetron/public/common/time.proto \
       spacex_api/telemetron/public/integrations/rate_limit_reason.proto \
       spacex_api/common/status/status.proto

echo "Proto generation complete."
