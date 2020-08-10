package umonitor

import (
	"reflect"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"go.uber.org/zap"
)

//DescribeDB

const typeUDB = "udb"

var (
	udbChan         = make(chan *resourceLabels, 500)
	udbResource     = udbListNew()
	typeUDBnamelist = []string{"nosql", "postgresql", "sql"}
)

// 初始化，注册udb功能函数
func init() {
	registerResource(typeUDB, udbResourceUpdate)
	registry.MustRegister(udbResource)
}

// udb函数
func udbResourceUpdate(uauth *UAuth, uzone *uZoneInfo, resourceMetric *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string) {

	if nil == resourceMetric {
		resourceMetric = new(ucloudResourceMetrics)
		resourceMetric.ResourceType = uMetricsNew(typeUDB)
		resourceMetric.ResourceIDList = make(map[string]*resourceLabels)
	}
	resourceMetric.Lock()
	defer resourceMetric.Unlock()
	resourceMetric.ResourceType.Upate(uauth)

	uclient := udb.NewClient(uauth.cfg, uauth.cre)
	num := len(uzone.projectsInfo) * len(uzone.regionInfo) * len(typeUDBnamelist)

	for projectID, projectName := range uzone.projectsInfo {
		for region := range uzone.regionInfo {
			go udbInstanceRequest(uclient, projectID, projectName, region)
		}
	}

	for {
		select {
		case resourcelabels := <-udbChan:
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
		zap.String("type", string(typeUDB)),
	)
	return resourceMetric, nil, "udbResourceUpdate"
}

func udbInstanceRequest(uclient *udb.UDBClient, projectID, projectName, region string) {

	offset := 0
	limit := 50

	udbReq := uclient.NewDescribeUDBInstanceRequest()
	udbReq.ProjectId = &projectID
	udbReq.Region = &region
	udbReq.Offset = &offset
	udbReq.Limit = &limit

	for _, udbTypeName := range typeUDBnamelist {
		udbReq.ClassType = &udbTypeName
		udbList, _ := uclient.DescribeUDBInstance(udbReq)
		if udbList.TotalCount == 0 {
			selfConf.logger.Debug("Not resource.",
				zap.String("Project", string(projectName)),
				zap.String("Region", string(region)),
				zap.String("Type", string(udbTypeName)),
			)
			udbChan <- nil
			continue
		}
		// 匿名函数
		selfFun := func(db interface{}, selfProjectID, selfProjectName, selfRegion, sqlType string) {
			switch db.(type) {
			case udb.UDBInstanceSet:
				masterResource(db.(udb.UDBInstanceSet), selfProjectID, selfProjectName, selfRegion, sqlType)
			case udb.UDBSlaveInstanceSet:
				slaveResource(db.(udb.UDBSlaveInstanceSet), selfProjectID, selfProjectName, selfRegion, sqlType)
			default:
				selfConf.logger.Warn("resource is not udb ",
					zap.Any("Type", reflect.TypeOf(db)),
				)
				return
			}
		}
		//
		for i := 0; i < udbList.TotalCount; i = i + limit {
			offset = i
			if offset > 0 {
				udbList, _ = uclient.DescribeUDBInstance(udbReq)
			}

			for _, udb := range udbList.DataSet {

				selfFun(udb, projectID, projectName, region, udbTypeName)
				for _, slaveDB := range udb.DataSet {
					selfFun(slaveDB, projectID, projectName, region, udbTypeName)
				}
			}
		}
	}
	udbChan <- nil
}

func masterResource(db udb.UDBInstanceSet, projectID, projectName, region, sqlType string) {
	resourcelabels := new(resourceLabels)
	resourcelabels.project_id = projectID
	resourcelabels.project_name = projectName
	resourcelabels.region_id = region
	resourcelabels.resource_type = typeUDB
	resourcelabels.resource_id = db.DBId
	resourcelabels.zone_id = db.Zone
	resourcelabels.resource_name = db.Name
	resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
	udbChan <- resourcelabels

	go func(selfresourcelabels *resourceLabels, selfDB udb.UDBInstanceSet, selfSqlType string) {
		udbResource.Lock()
		defer udbResource.Unlock()
		_, found := udbResource.labels[selfresourcelabels.resource_id]
		if !found {
			udbResource.labels[resourcelabels.resource_id] = new(udbLables)
		}
		udbResource.labels[resourcelabels.resource_id].baseLables = selfresourcelabels
		udbResource.labels[resourcelabels.resource_id].instance_mode = selfDB.InstanceMode
		udbResource.labels[resourcelabels.resource_id].instance_type = selfDB.InstanceType
		udbResource.labels[resourcelabels.resource_id].sql_type = selfSqlType
		udbResource.labels[resourcelabels.resource_id].src_db_id = selfDB.SrcDBId
		udbResource.labels[resourcelabels.resource_id].virtual_ip = selfDB.VirtualIP
		udbResource.labels[resourcelabels.resource_id].role = selfDB.Role
		udbResource.labels[resourcelabels.resource_id].status = selfDB.State
		udbResource.labels[resourcelabels.resource_id].tag = selfDB.Tag

	}(resourcelabels, db, sqlType)

}

func slaveResource(db udb.UDBSlaveInstanceSet, projectID, projectName, region, sqlType string) {
	resourcelabels := new(resourceLabels)
	resourcelabels.project_id = projectID
	resourcelabels.project_name = projectName
	resourcelabels.region_id = region
	resourcelabels.resource_type = typeUDB
	resourcelabels.resource_id = db.DBId
	resourcelabels.zone_id = db.Zone
	resourcelabels.resource_name = db.Name
	resourcelabels.hashid = resourcelabels.resource_id + resourcelabels.project_id + resourcelabels.region_id + resourcelabels.resource_name
	udbChan <- resourcelabels

	go func(selfresourcelabels *resourceLabels, selfDB udb.UDBSlaveInstanceSet, selfSqlType string) {
		udbResource.Lock()
		defer udbResource.Unlock()
		_, found := udbResource.labels[selfresourcelabels.resource_id]
		if !found {
			udbResource.labels[resourcelabels.resource_id] = new(udbLables)
		}
		udbResource.labels[resourcelabels.resource_id].baseLables = selfresourcelabels
		udbResource.labels[resourcelabels.resource_id].instance_mode = selfDB.InstanceMode
		udbResource.labels[resourcelabels.resource_id].instance_type = selfDB.InstanceType
		udbResource.labels[resourcelabels.resource_id].sql_type = selfSqlType
		udbResource.labels[resourcelabels.resource_id].src_db_id = selfDB.SrcDBId
		udbResource.labels[resourcelabels.resource_id].virtual_ip = selfDB.VirtualIP
		udbResource.labels[resourcelabels.resource_id].role = selfDB.Role
		udbResource.labels[resourcelabels.resource_id].status = selfDB.State
		udbResource.labels[resourcelabels.resource_id].tag = selfDB.Tag

	}(resourcelabels, db, sqlType)
}

// udb resource list
type udbList struct {
	labels map[string]*udbLables
	sync.RWMutex
}

type udbLables struct {
	baseLables    *resourceLabels
	sql_type      string
	src_db_id     string
	instance_type string
	instance_mode string
	virtual_ip    string
	role          string
	status        string
	tag           string
}

func udbListNew() *udbList {
	return &udbList{
		labels: make(map[string]*udbLables, 0),
	}
}

func (udb *udbList) Describe(ch chan<- *prometheus.Desc) {
	for range udb.labels {
		ch <- prometheus.NewDesc(
			"ucloud_udb_info",
			"ucloud udb info and labels",
			[]string{
				"project_id",
				"project_name",
				"region_id",
				"resource_name",
				"resource_id",
				"sql_type",
				"src_db_id",
				"instance_type",
				"instance_mode",
				"virtual_ip",
				"role",
				"status",
				"tag",
			},
			nil,
		)
	}
}

func (udb *udbList) Collect(ch chan<- prometheus.Metric) {

	for _, udbLable := range udb.labels {

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"ucloud_udb_info",
				"ucloud udb info and labels",
				[]string{
					"project_id",
					"project_name",
					"region_id",
					"resource_name",
					"resource_id",
					"sql_type",
					"src_db_id",
					"instance_type",
					"instance_mode",
					"virtual_ip",
					"role",
					"status",
					"tag",
				},
				nil,
			),
			prometheus.GaugeValue,
			1,
			[]string{
				udbLable.baseLables.project_id,
				udbLable.baseLables.project_name,
				udbLable.baseLables.region_id,
				udbLable.baseLables.resource_name,
				udbLable.baseLables.resource_id,
				udbLable.sql_type,
				udbLable.src_db_id,
				udbLable.instance_type,
				udbLable.instance_mode,
				udbLable.virtual_ip,
				udbLable.role,
				udbLable.status,
				udbLable.tag,
			}...,
		)

	}
}
