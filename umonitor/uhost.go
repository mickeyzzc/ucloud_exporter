package umonitor

import (
	"runtime"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"go.uber.org/zap"
)

const typeUHost = "uhost"

var (
	uhostChan = make(chan *resourceLabels, 1000)
)

func init() {
	registerResource(typeUHost, uhostResourceUpdate)
}

func uhostResourceUpdate(uauth *UAuth, uzone *uZoneInfo, resourceMetric *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string) {
	selfFunc, _, _, _ := runtime.Caller(1)
	if nil == resourceMetric {
		resourceMetric = new(ucloudResourceMetrics)
		resourceMetric.ResourceType = uMetricsNew(typeUHost)
		resourceMetric.ResourceIDList = make(map[string]*resourceLabels)
	}
	resourceMetric.Lock()
	defer resourceMetric.Unlock()
	resourceMetric.ResourceType.Upate(uauth)

	uclient := uhost.NewClient(uauth.cfg, uauth.cre)
	num := len(uzone.projectsInfo) * len(uzone.regionInfo)

	for projectID, projectName := range uzone.projectsInfo {
		for region := range uzone.regionInfo {
			go uHostInstanceRequest(uclient, projectID, projectName, region)
		}
	}

	for {
		select {
		case resourcelabels := <-uhostChan:
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
		zap.String("type", string(typeUHost)),
	)
	return resourceMetric, nil, runtime.FuncForPC(selfFunc).Name()
}

func uHostInstanceRequest(uclient *uhost.UHostClient, projectID, projectName, region string) {

	offset := 0
	limit := 200

	uhostReq := uclient.NewDescribeUHostInstanceRequest()
	uhostReq.ProjectId = &projectID
	uhostReq.Region = &region
	uhostReq.Offset = &offset
	uhostReq.Limit = &limit

	uHostList, _ := uclient.DescribeUHostInstance(uhostReq)
	if uHostList.TotalCount == 0 {
		selfConf.logger.Debug("Not resource.",
			zap.String("Project", string(projectName)),
			zap.String("Region", string(region)),
			zap.String("Type", string(typeUHost)),
		)
		uhostChan <- nil
		return
	}
	selfConf.logger.Info("",
		zap.String("Project", string(projectName)),
		zap.String("Region", string(region)),
		zap.String("Type", string(typeUHost)),
		zap.Int("resource_num", uHostList.TotalCount),
	)
	for i := 0; i < uHostList.TotalCount; i = i + limit {
		offset = i
		if offset > 0 {
			uHostList, _ = uclient.DescribeUHostInstance(uhostReq)
		}

		for _, host := range uHostList.UHostSet {
			resourcelabels := new(resourceLabels)
			resourcelabels.project_id = projectID
			resourcelabels.project_name = projectName
			resourcelabels.region_id = region
			resourcelabels.resource_type = typeUHost
			resourcelabels.resource_id = host.UHostId
			resourcelabels.zone_id = host.Zone
			resourcelabels.resource_name = host.Name
			resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
			uhostChan <- resourcelabels
		}
	}
	uhostChan <- nil
}
