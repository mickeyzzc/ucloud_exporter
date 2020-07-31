package umonitor

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"go.uber.org/zap"
)

//DescribeEIP
const typeEIP = "eip"

var (
	ueipChan    = make(chan *resourceLabels, 500)
	eipResource = eipListNew()
)

// 初始化，注册ueip功能函数
func init() {
	registerResource(typeEIP, ueipResourceUpdate)
	registry.MustRegister(eipResource)
}

// ueip函数
func ueipResourceUpdate(uauth *UAuth, uzone *uZoneInfo, resourceMetric *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string) {

	if nil == resourceMetric {
		resourceMetric = new(ucloudResourceMetrics)
		resourceMetric.ResourceType = uMetricsNew(typeEIP)
		resourceMetric.ResourceIDList = make(map[string]*resourceLabels)
	}
	resourceMetric.Lock()
	defer resourceMetric.Unlock()
	resourceMetric.ResourceType.Upate(uauth)

	uclient := unet.NewClient(uauth.cfg, uauth.cre)
	num := len(uzone.projectsInfo) * len(uzone.regionInfo)

	for projectID, projectName := range uzone.projectsInfo {
		for region := range uzone.regionInfo {
			go ueipInstanceRequest(uclient, projectID, projectName, region)
		}
	}

	for {
		select {
		case resourcelabels := <-ueipChan:
			if nil == resourcelabels {
				num = num - 1
				if num == 0 {
					goto ForEnd
				}
				continue
			}
			rlabel, found := resourceMetric.ResourceIDList[resourcelabels.resource_id]
			hashid := resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
			if found {
				if rlabel.hashid == hashid {
					continue
				}
			}
			selfConf.logger.Debug("update resource",
				zap.String("id", string(resourcelabels.resource_id)),
				zap.String("hashid", string(hashid)),
			)
			resourceMetric.ResourceIDList[resourcelabels.resource_id] = resourcelabels
		}
	}

ForEnd:
	selfConf.logger.Info(
		"resource update ok",
		zap.String("type", string(typeEIP)),
	)
	return resourceMetric, nil, "ueipResourceUpdate"
}

func ueipInstanceRequest(uclient *unet.UNetClient, projectID, projectName, region string) {

	offset := 0
	limit := 50

	ueipReq := uclient.NewDescribeEIPRequest()
	ueipReq.ProjectId = &projectID
	ueipReq.Region = &region
	ueipReq.Offset = &offset
	ueipReq.Limit = &limit
	selfZone := "none"
	ueipList, _ := uclient.DescribeEIP(ueipReq)
	if ueipList.TotalCount == 0 {
		selfConf.logger.Debug("Not resource.",
			zap.String("Project", string(projectName)),
			zap.String("Region", string(region)),
			zap.String("Type", string(typeEIP)),
		)
		ueipChan <- nil
		return
	}
	selfConf.logger.Info("",
		zap.String("Project", string(projectName)),
		zap.String("Region", string(region)),
		zap.String("Type", string(typeEIP)),
		zap.Int("resource_num", ueipList.TotalCount),
	)
	for i := 0; i < ueipList.TotalCount; i = i + limit {
		offset = i
		if offset > 0 {
			ueipList, _ = uclient.DescribeEIP(ueipReq)
		}
		for _, eip := range ueipList.EIPSet {
			resourcelabels := new(resourceLabels)
			resourcelabels.project_id = projectID
			resourcelabels.project_name = projectName
			resourcelabels.region_id = region
			resourcelabels.resource_type = typeEIP
			resourcelabels.resource_id = eip.EIPId
			resourcelabels.zone_id = selfZone
			resourcelabels.resource_name = eip.Name
			resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
			ueipChan <- resourcelabels
			//
			go func(selfresourcelabels *resourceLabels, selfeip unet.UnetEIPSet) {
				eipResource.Lock()
				defer eipResource.Unlock()
				_, found := eipResource.labels[selfresourcelabels.resource_id]
				if !found {
					eipResource.labels[resourcelabels.resource_id] = new(eipLables)
				}
				eipResource.labels[resourcelabels.resource_id].baseLables = selfresourcelabels
				eipResource.labels[resourcelabels.resource_id].eip_ipaddr = selfeip.EIPAddr[0].IP
				eipResource.labels[resourcelabels.resource_id].eip_operator = selfeip.EIPAddr[0].OperatorName
				eipResource.labels[resourcelabels.resource_id].bind_resource_type = selfeip.Resource.ResourceType
				eipResource.labels[resourcelabels.resource_id].bind_resource_id = selfeip.Resource.ResourceId
				eipResource.labels[resourcelabels.resource_id].bind_resource_name = selfeip.Resource.ResourceName
				eipResource.labels[resourcelabels.resource_id].status = selfeip.Status
				eipResource.labels[resourcelabels.resource_id].tag = selfeip.Tag
				eipResource.labels[resourcelabels.resource_id].bandwidth = Interface2String(selfeip.Bandwidth)
			}(resourcelabels, eip)

		}
	}
	ueipChan <- nil
}

type eipList struct {
	labels map[string]*eipLables
	sync.RWMutex
}

type eipLables struct {
	baseLables         *resourceLabels
	eip_ipaddr         string
	eip_operator       string
	bind_resource_type string
	bind_resource_id   string
	bind_resource_name string
	status             string
	tag                string
	bandwidth          string
}

func eipListNew() *eipList {
	return &eipList{
		labels: make(map[string]*eipLables, 0),
	}
}

func (eip *eipList) Describe(ch chan<- *prometheus.Desc) {
	for range eip.labels {
		ch <- prometheus.NewDesc(
			"ucloud_eip_info",
			"ucloud eip info and labels",
			[]string{
				"project_id",
				"project_name",
				"region_id",
				"resource_name",
				"resource_id",
				"eip_ipaddr",
				"eip_operator",
				"bind_resource_type",
				"bind_resource_id",
				"bind_resource_name",
				"status",
				"tag",
				"bandwidth",
			},
			nil,
		)
	}
}

func (eip *eipList) Collect(ch chan<- prometheus.Metric) {

	for _, eipLable := range eip.labels {

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"ucloud_eip_info",
				"ucloud eip info and labels",
				[]string{
					"project_id",
					"project_name",
					"region_id",
					"resource_name",
					"resource_id",
					"eip_ipaddr",
					"eip_operator",
					"bind_resource_type",
					"bind_resource_id",
					"bind_resource_name",
					"status",
					"tag",
					"bandwidth",
				},
				nil,
			),
			prometheus.GaugeValue,
			1,
			[]string{
				eipLable.baseLables.project_id,
				eipLable.baseLables.project_name,
				eipLable.baseLables.region_id,
				eipLable.baseLables.resource_name,
				eipLable.baseLables.resource_id,
				eipLable.eip_ipaddr,
				eipLable.eip_operator,
				eipLable.bind_resource_type,
				eipLable.bind_resource_id,
				eipLable.bind_resource_name,
				eipLable.status,
				eipLable.tag,
				eipLable.bandwidth,
			}...,
		)

	}
}
