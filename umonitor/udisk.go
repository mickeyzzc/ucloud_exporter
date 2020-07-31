package umonitor

import (
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"go.uber.org/zap"
)

const typeUdisk = "udisk"

var (
	udiskChan = make(chan *resourceLabels, 1000)
)

// 初始化，注册udisk功能函数
func init() {
	registerResource(typeUdisk, udiskResourceUpdata)
}

// udisk函数
func udiskResourceUpdata(uauth *UAuth, uzone *uZoneInfo, resourceMetric *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string) {

	if nil == resourceMetric {
		resourceMetric = new(ucloudResourceMetrics)
		resourceMetric.ResourceType = uMetricsNew(typeUdisk)
		resourceMetric.ResourceIDList = make(map[string]*resourceLabels)
	}
	resourceMetric.Lock()
	defer resourceMetric.Unlock()
	resourceMetric.ResourceType.Upate(uauth)

	uclient := udisk.NewClient(uauth.cfg, uauth.cre)
	num := len(uzone.projectsInfo) * len(uzone.regionInfo)

	for projectID, projectName := range uzone.projectsInfo {
		for region := range uzone.regionInfo {
			go udiskInstanceRequest(uclient, projectID, projectName, region)
		}
	}

	for {
		select {
		case resourcelabels := <-udiskChan:
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
		zap.String("type", string(typeUdisk)),
	)
	return resourceMetric, nil, "udiskResourceUpdata"
}

func udiskInstanceRequest(uclient *udisk.UDiskClient, projectID, projectName, region string) {

	offset := 0
	limit := 300

	udiskReq := uclient.NewDescribeUDiskRequest()
	udiskReq.ProjectId = &projectID
	udiskReq.Region = &region
	udiskReq.Offset = &offset
	udiskReq.Limit = &limit

	udiskList, _ := uclient.DescribeUDisk(udiskReq)
	if udiskList.TotalCount == 0 {
		selfConf.logger.Debug("Not resource.",
			zap.String("Project", string(projectName)),
			zap.String("Region", string(region)),
			zap.String("Type", string(typeUdisk)),
		)
		udiskChan <- nil
		return
	}
	selfConf.logger.Info("",
		zap.String("Project", string(projectName)),
		zap.String("Region", string(region)),
		zap.String("Type", string(typeUdisk)),
		zap.Int("resource_num", udiskList.TotalCount),
	)
	for i := 0; i < udiskList.TotalCount; i = i + limit {
		offset = i
		if offset > 0 {
			udiskList, _ = uclient.DescribeUDisk(udiskReq)
		}

		for _, disk := range udiskList.DataSet {
			/*
			 ProtocolVersion字段为1时，需结合IsBoot确定具体磁盘类型:
			 普通数据盘：DiskType:"CLOUDNORMAL",IsBoot:"False"；
			 普通系统盘：DiskType:"CLOUDNORMAL",IsBoot:"True"；
			 SSD数据盘：DiskType:"CLOUDSSD",IsBoot:"False"；
			 SSD系统盘：DiskType:"CLOUDSSD",IsBoot:"True"；
			 RSSD数据盘：DiskType:"CLOUD_RSSD",IsBoot:"False"；
			 为空拉取所有。
			 ProtocolVersion字段为0或没有该字段时，可设为以下几个值:
			 普通数据盘：DataDisk； udisk
			 普通系统盘：SystemDisk；udisk_sys
			 SSD数据盘：SSDDataDisk；udisk_ssd
			 SSD系统盘：SSDSystemDisk；udisk_sys
			 RSSD数据盘：RSSDDataDisk； udisk_rssd
			 为空拉取所有。
			*/
			resourcelabels := new(resourceLabels)
			resourcelabels.project_id = projectID
			resourcelabels.project_name = projectName
			resourcelabels.region_id = region
			switch disk.DiskType {
			case "RSSDDataDisk":
				resourcelabels.resource_type = "udisk_rssd"
			case "SSDDataDisk":
				resourcelabels.resource_type = "udisk_ssd"
			case "DataDisk":
				resourcelabels.resource_type = "udisk"
			default:
				resourcelabels.resource_type = "udisk_sys"
			}
			resourcelabels.resource_id = disk.UDiskId
			resourcelabels.zone_id = disk.Zone
			resourcelabels.resource_name = disk.Name
			resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
			udiskChan <- resourcelabels
		}
	}
	udiskChan <- nil
}
