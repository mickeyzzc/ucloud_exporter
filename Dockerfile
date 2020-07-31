From golang:1.13.1-alpine3.10 as builder

RUN apk add --no-cache \
  wget \
  git 

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
  && chmod +x /usr/local/bin/dumb-init

RUN mkdir -p /go/src/gitee.com/MickeyZZC/ucloud_monitor_api
WORKDIR /go/src/gitee.com/MickeyZZC/ucloud_monitor_api

# Cache dependencies
COPY go.mod .
COPY go.sum .

RUN export GOPROXY=https://goproxy.io && GO111MODULE=on go mod download

# Build real binaries
COPY . .
RUN go build

# Executable image
FROM alpine

COPY --from=builder /go/src/gitee.com/MickeyZZC/ucloud_exporter/ucloud_exporter /ucloud_exporter
COPY --from=builder /usr/local/bin/dumb-init /usr/local/bin/dumb-init

WORKDIR /

ENTRYPOINT ["/usr/local/bin/dumb-init", "/ucloud_exporter"]

