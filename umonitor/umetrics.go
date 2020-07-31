package umonitor

import (
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/private/services/umon"
	"go.uber.org/zap"
)

var (
	labelsIndex = []string{
		"project_id",
		"project_name",
		"region_id",
		"zone_id",
		"resource_name",
		"resource_id",
		"resource_type",
	}
)

type ucloudMetrics struct {
	ResourceType        string
	ResourceMetricNames *[]string
	ResourceMetrics     map[string]*uResourceMetric
	sync.RWMutex
}

type uResourceMetric struct {
	Type      string
	Unit      string
	Frequency int
	Desc      *prometheus.Desc
}

func uMetricsNew(typeName string) *ucloudMetrics {
	return &ucloudMetrics{
		ResourceType:    typeName,
		ResourceMetrics: make(map[string]*uResourceMetric),
	}
}

// 获取监控项目名称
func (ums *ucloudMetrics) Upate(uauth *UAuth) error {
	uClient := umon.NewClient(uauth.cfg, uauth.cre)
	rMetricsRequest := uClient.NewDescribeResourceMetricRequest()
	rMetricsRequest.ResourceType = &ums.ResourceType

	rMetricsResponse, err := uClient.DescribeResourceMetric(rMetricsRequest)
	if err != nil {
		return err
	}
	resourceMetrics := ums.ResourceMetrics
	resourceMetricsKey := make([]string, 0)
	for i := 0; i < len(rMetricsResponse.DataSet); i++ {
		metricName := rMetricsResponse.DataSet[i].MetricName
		metric, found := resourceMetrics[metricName]
		if !found {
			metric = new(uResourceMetric)
			resourceMetrics[metricName] = metric
		}
		resourceMetricsKey = append(resourceMetricsKey, metricName)
		metric.Frequency = rMetricsResponse.DataSet[i].Frequency
		metric.Type = rMetricsResponse.DataSet[i].Type
		metric.Unit = rMetricsResponse.DataSet[i].Unit
	}
	ums.ResourceMetricNames = &resourceMetricsKey
	ums.ResourceMetrics = resourceMetrics
	ums.umsDescUpdate()
	return nil
}

func (ums *ucloudMetrics) umsDescUpdate() {

	for umetrics, umetricsInfo := range ums.ResourceMetrics {
		metricName := strings.Join(
			[]string{
				ums.ResourceType,
				umetrics,
			}, "_",
		)
		docString := strings.Join(
			[]string{
				"Resource type is",
				ums.ResourceType,
				", Metric name is",
				umetrics,
				", unit is",
				umetricsInfo.Unit,
				", frequency is",
				Interface2String(umetricsInfo.Frequency),
			}, " ",
		)

		desc := prometheus.NewDesc(
			metricName,
			docString,
			labelsIndex,
			nil,
		)

		umetricsInfo.Desc = desc

	}

}

type uMetricsValue struct {
	value map[time.Time]float64
}

// 调API获取监控数据
func (ums *ucloudMetrics) GetValue(uClient *umon.UMonClient, project, region, zone, resourceID, resourceType string, timeRange, timeBegin, timeEnd int) (*map[string]*uMetricsValue, error) {
	defer func() {
		if err := recover(); err != nil {
			selfConf.logger.Error("func get metrics err",
				zap.Any("msg", err),
			)
		}
	}()

	req := uClient.NewGetMetricRequest()

	req.ProjectId = &project
	req.Region = &region
	if len(zone) > 0 {
		if zone != "none" {
			req.Zone = &zone
		}

	}
	req.ResourceId = &resourceID
	req.ResourceType = &ums.ResourceType
	if len(resourceType) != 0 {
		req.ResourceType = &resourceType
	}
	// 2009-08-11 22:13:20
	if timeBegin > 1250000000 && timeEnd > timeBegin {
		req.BeginTime = &timeBegin
		req.EndTime = &timeEnd
	} else {
		req.TimeRange = &timeRange
	}

	req.MetricName = *ums.ResourceMetricNames

	resp, err := uClient.GetMetric(req)
	if err != nil {
		selfConf.logger.Warn(
			"get value err ",
			zap.Any("id", &req.ResourceId),
			zap.Any("typeName", &req.ResourceType),
			zap.Any("labels.project_id", &req.ProjectId),
			zap.Any("labels.region_id", &req.Region),
			zap.Any("labels.zone_id", &req.Zone),
			zap.Any("MetricName", &req.MetricName),
		)
		return nil, err
	}
	metricsLists := make(map[string]*uMetricsValue, 0)

	for name, values := range resp.DataSets {
		metricsValue := new(uMetricsValue)
		timeIndex := make(map[time.Time]float64)
		if nil == values {
			continue
		}
		if len(values) == 0 {
			continue
		}
		for _, metric := range values {
			timestamp := time.Unix(int64(metric.Timestamp), 0)
			f64, err := interface2Float64(metric.Value)
			if nil != err {
				continue
			}
			selfConf.logger.Debug(
				"get value from api",
				zap.Any("id", &req.ResourceId),
				zap.Any("typeName", name),
				zap.Any("labels.project_id", &req.ProjectId),
				zap.Any("labels.region_id", &req.Region),
				zap.Any("labels.zone_id", &req.Zone),
				zap.Any("timestamp", timestamp),
				zap.Any("value", metric.Value),
			)
			timeIndex[timestamp] = f64
		}
		metricsValue.value = timeIndex
		metricsLists[name] = metricsValue
	}
	return &metricsLists, nil
}
