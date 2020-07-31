package ukafka

/*
MetricItem - GetMetricInfo-监控项信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type MetricItem struct {
	Value interface{}

	Timestamp    int
	DistrictName string
	IspName      string
}

type MetricItemSet struct {
	MetricName  string
	Granularity int
	MetricData  []MetricItem
}
