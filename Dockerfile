FROM golang:1.25 AS build
WORKDIR /go/src/app
ENV GO111MODULE=auto CGO_ENABLED=0
COPY . .
RUN go build

FROM alpine:latest
WORKDIR /app
COPY --from=build /go/src/app /app
ENTRYPOINT [ "./starlink-exporter" ]
