package umonitor

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ucloud/ucloud-sdk-go/private/services/umon"
	"go.uber.org/zap"
)

var registry = prometheus.NewRegistry()

func init() {
	registry.MustRegister(selfResource)
	registry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
}

func PrometheusColletcor(metricsPath string, listenAddr int) {

	http.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Ucloud Exporter</title></head>
		<body>
		<h1>Ucloud Exporter</h1>
		<p><a href="` + metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})

	selfConf.logger.Info(
		"Starting Server at ",
		zap.String("listen", strconv.Itoa(listenAddr)),
		zap.String("path", metricsPath),
	)
	http.ListenAndServe(":"+strconv.Itoa(listenAddr), nil)
}

// Describe implements the prometheus.Collector interface.
func (ucrs *ucloudResources) Describe(ch chan<- *prometheus.Desc) {
	for _, resourcs := range ucrs.resourceList {
		for _, rMetric := range resourcs.ResourceType.ResourceMetrics {
			ch <- rMetric.Desc
		}
	}
}

func (ucrs *ucloudResources) Collect(ch chan<- prometheus.Metric) {
	ucrs.RLock() // 加锁
	defer ucrs.RUnlock()

	defer func() {
		selfConf.logger.Sync()

		if err := recover(); err != nil {
			selfConf.logger.Error("Collect err",
				zap.Any("msg", err),
			)
		}
	}()

	var pool = PoolNew(50)
	uClient := umon.NewClient(selfConf.uauth.cfg, selfConf.uauth.cre)
	var tmp_num = 0
	selfConf.logger.Info("ucloud metrics collect now")
	for nameTmp, resource := range ucrs.resourceList {

		resource.RLock()
		defer resource.RUnlock()
		selfConf.logger.Debug(
			"collectoring ",
			zap.String("nameTmp", nameTmp),
			zap.Any("resource", resource.ResourceType.ResourceType),
		)
		metricTypeList := resource.ResourceType.ResourceMetrics

		for id, labels := range resource.ResourceIDList {
			selfConf.logger.Debug(
				"try to get umon value",
				zap.Any("project", labels.project_id),
				zap.Any("region", labels.region_id),
				zap.Any("hashid", labels.hashid),
				zap.Any("resource_id", id),
			)

			pool.Add(1)
			go func(resourceType *ucloudMetrics, selfID string, selfLabels *resourceLabels) {
				defer pool.Done()
				metricsValues, err := resourceType.GetValue(uClient, selfLabels.project_id, selfLabels.region_id, selfLabels.zone_id, selfID, selfLabels.resource_type, 600, 0, 0)
				if nil != err {
					selfConf.logger.Warn(
						"get umon value err",
						zap.String("project", selfLabels.project_id),
						zap.String("region", selfLabels.region_id),
						zap.String("resource_id", selfID),
						zap.String("resource_type", selfLabels.resource_type),
					)
					return
				}
				metricLabels := []string{
					selfLabels.project_id,
					selfLabels.project_name,
					selfLabels.region_id,
					selfLabels.zone_id,
					selfLabels.resource_name,
					selfLabels.resource_id,
					selfLabels.resource_type,
				}
				tmp_num = tmp_num + 1
				for typeName, values := range *metricsValues {
					metricType, found := metricTypeList[typeName]
					if !found {
						continue
					}
					tmp_timestamp := time.Unix(int64(0), 0)
					var tmp_value float64 = 0.0
					for timestamp, value := range values.value {
						if timestamp.After(tmp_timestamp) {
							tmp_timestamp = timestamp
							tmp_value = value
						}
					}
					if tmp_timestamp.After(time.Unix(int64(1), 0)) {
						selfConf.logger.Debug(
							"collector value ",
							zap.Any("id", selfID),
							zap.Any("typeName", typeName),
							zap.Any("labels.project_id", selfLabels.project_id),
							zap.Any("labels.region_id", selfLabels.region_id),
							zap.Any("labels.zone_id", selfLabels.zone_id),
							zap.Any("labels.resource_name", selfLabels.resource_name),
							zap.Any("timestamp", tmp_timestamp),
							zap.Any("value", tmp_value),
						)
						ch <- prometheus.NewMetricWithTimestamp(
							tmp_timestamp,
							prometheus.MustNewConstMetric(
								metricType.Desc,
								prometheus.GaugeValue,
								tmp_value,
								metricLabels...,
							),
						)
					}

				}

			}(resource.ResourceType, id, labels)

		}
	}
	pool.Wait()
	selfConf.logger.Info("ucloud metrics collect over", zap.Int("collect_num", tmp_num))
}
