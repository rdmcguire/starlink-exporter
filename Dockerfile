FROM golang:1.27 AS build
WORKDIR /go/src/app
ENV GO111MODULE=auto CGO_ENABLED=0
COPY . .
RUN go build -o starlink-exporter ./cmd/starlink-exporter

FROM alpine:latest
WORKDIR /app
COPY --from=build /go/src/app/starlink-exporter /app/
ENTRYPOINT [ "./starlink-exporter" ]
