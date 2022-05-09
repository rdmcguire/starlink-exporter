#/bin/bash

protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/device.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/common/status/status.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/command.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/common.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/dish.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/dish_config.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/wifi.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/wifi_config.proto
protoc --go_out=.    --go-grpc_out=.  --descriptor_set_in=proto/dish.protoset spacex/api/device/transceiver.proto
