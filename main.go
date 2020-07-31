package main

import (
	"flag"

	"gitee.com/mickeybee/ucloud_exporter/umonitor"
	"go.uber.org/zap"
)

var (
	// 命令行参数
	ucloudPrivateKey   = flag.String("ucloud.privatekey", "", "Ucloud api  PrivateKey.")
	ucloudPublicKey    = flag.String("ucloud.publickey", "", "Ucloud api  PublicKey.")
	ucloudTimeInterval = flag.Int64("ucloud.interval", 3600, "Ucloud time interval. ")
	logPath            = flag.String("log.path", "./ucloud_exporter.log", "log path")
	logLevel           = flag.String("log.level", "info", "'info','debug'")
	// 命令行参数
	listenAddr  = flag.Int("web.listen-port", 58086, "An port to listen on for web interface and telemetry.")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
)

func main() {
	flag.Parse()
	logger := umonitor.InitLogger(*logPath, *logLevel)
	defer logger.Sync()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("err", zap.Any("msg", err))
		}
	}()

	umonitor.InitConf(
		*ucloudPrivateKey,
		*ucloudPublicKey,
		nil,
		logger,
	)

	go umonitor.ResourceHandle(ucloudTimeInterval)
	umonitor.PrometheusColletcor(*metricsPath, *listenAddr)
}
