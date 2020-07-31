package usocketmon

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

/*
https://api.ucloud.cn/?ProjectId=org-12838&Region=cn-bj2&DistrictType=china&TimeRange=3600&ResourceType=eip&PublicIP.0=123.59.78.221&ResourceId.0=eip-hnvfq3&MetricName.0=InUDPPkts&MetricName.1=OutTCPPkts&Action=GetShockwaveMetric

{
  "RetCode": 0,
  "Action": "GetShockwaveMetricResponse",
  "DataSet": [
    {
      "MetricName": "InUDPPkts",
      "Granularity": 3600,
      "MetricData": [
        {
          "Timestamp": 1566208800,
          "Value": 0,
          "DistrictName": "Total",
          "IspName": "Total"
        }
      ]
    },
    {
      "MetricName": "OutUDPPkts",
      "Granularity": 3600,
      "MetricData": [
        {
          "Timestamp": 1566208800,
          "Value": 0,
          "DistrictName": "Total",
          "IspName": "Total"
        }
      ]
    }
  ]
}
*/
