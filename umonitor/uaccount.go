package umonitor

import (
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UAuth struct {
	cfg *ucloud.Config
	cre *auth.Credential
}

func uAuthNew(pri, pub string) *UAuth {
	cfg := ucloud.NewConfig()
	credential := auth.NewCredential()
	credential.PrivateKey = pri
	credential.PublicKey = pub
	return &UAuth{
		cfg: &cfg,
		cre: &credential,
	}
}

func (ua *UAuth) GetUauthInfo() (*ucloud.Config, *auth.Credential) {
	return ua.cfg, ua.cre
}

type uZoneInfo struct {
	projectsInfo map[string]string
	regionInfo   map[string]*[]string
}

// 获取账号授权范围基本信息
func (ua *UAuth) GetBaseAccountZoneList() (*uZoneInfo, error) {

	uaccountClient := uaccount.NewClient(ua.cfg, ua.cre)

	projectRequest := uaccountClient.NewGetProjectListRequest()
	projectList, err := uaccountClient.GetProjectList(projectRequest)
	if err != nil {
		return nil, err
	}

	regionRequest := uaccountClient.NewGetRegionRequest()
	regionList, err := uaccountClient.GetRegion(regionRequest)
	if err != nil {
		return nil, err
	}

	projectInfo := make(map[string]string)
	for i := 0; i < projectList.ProjectCount; i++ {
		projectInfo[projectList.ProjectSet[i].ProjectId] = projectList.ProjectSet[i].ProjectName
	}

	regionInfo := make(map[string]*[]string)
	for i := 0; i < len(regionList.Regions); i++ {
		region := regionList.Regions[i].Region
		zone := regionList.Regions[i].Zone

		zones, found := regionInfo[region]
		if found {
			*zones = append(*zones, zone)
			continue
		}
		newZones := []string{zone}
		zones = &newZones
		regionInfo[region] = zones
	}

	return &uZoneInfo{
		projectsInfo: projectInfo,
		regionInfo:   regionInfo,
	}, nil
}

//
