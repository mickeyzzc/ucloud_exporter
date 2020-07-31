package umonitor

import (
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"go.uber.org/zap"
)

//DescribeDB
const typeURedis = "uredis"

var (
	uredisChan = make(chan *resourceLabels, 500)
)

// 初始化，注册uredis功能函数
func init() {
	registerResource(typeURedis, uredisResourceUpdate)
}

// uredis函数
func uredisResourceUpdate(uauth *UAuth, uzone *uZoneInfo, resourceMetric *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string) {

	if nil == resourceMetric {
		resourceMetric = new(ucloudResourceMetrics)
		resourceMetric.ResourceType = uMetricsNew(typeURedis)
		resourceMetric.ResourceIDList = make(map[string]*resourceLabels)
	}
	resourceMetric.Lock()
	defer resourceMetric.Unlock()
	resourceMetric.ResourceType.Upate(uauth)

	uclient := umem.NewClient(uauth.cfg, uauth.cre)
	num := len(uzone.projectsInfo) * len(uzone.regionInfo)

	for projectID, projectName := range uzone.projectsInfo {
		for region := range uzone.regionInfo {
			go uredisInstanceRequest(uclient, projectID, projectName, region)
		}
	}

	for {
		select {
		case resourcelabels := <-uredisChan:
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
		zap.String("type", string(typeURedis)),
	)
	return resourceMetric, nil, "uredisResourceUpdate"
}

func uredisInstanceRequest(uclient *umem.UMemClient, projectID, projectName, region string) {

	offset := 0
	limit := 50

	umemReq := uclient.NewDescribeURedisGroupRequest()
	umemReq.ProjectId = &projectID
	umemReq.Region = &region
	umemReq.Offset = &offset
	umemReq.Limit = &limit
	umemList, _ := uclient.DescribeURedisGroup(umemReq)
	if umemList.TotalCount == 0 {
		selfConf.logger.Debug("Not resource.",
			zap.String("Project", string(projectName)),
			zap.String("Region", string(region)),
			zap.String("Type", string(typeURedis)),
		)
		uredisChan <- nil
		return
	}
	selfConf.logger.Info("",
		zap.String("Project", string(projectName)),
		zap.String("Region", string(region)),
		zap.String("Type", string(typeURedis)),
		zap.Int("resource_num", umemList.TotalCount),
	)
	for i := 0; i < umemList.TotalCount; i = i + limit {
		offset = i
		if offset > 0 {
			umemList, _ = uclient.DescribeURedisGroup(umemReq)
		}
		for _, umen := range umemList.DataSet {
			resourcelabels := new(resourceLabels)
			resourcelabels.project_id = projectID
			resourcelabels.project_name = projectName
			resourcelabels.region_id = region
			resourcelabels.resource_type = typeURedis
			resourcelabels.resource_id = umen.GroupId
			resourcelabels.zone_id = umen.Zone
			resourcelabels.resource_name = umen.Name
			resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
			uredisChan <- resourcelabels
		}
	}
	uredisChan <- nil
}
