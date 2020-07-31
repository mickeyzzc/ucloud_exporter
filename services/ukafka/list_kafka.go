package ukafka

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// DescribeULBRequest is request schema for DescribeULB action
type ListUKafkaInstance struct {
	request.CommonBase

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"true"`

	// 数据偏移量，默认为0
	Offset *int `required:"false"`

	// 数据分页值，默认为20
	Limit *int `required:"false"`

	ClusterInstanceId *string `required:"false"`

	Filter *string `required:"false"`
	VPCId  *string `required:"false"`

	SubnetId *string `required:"false"`

	BusinessId *string `required:"false"`
}

type ListUKafkaInstanceResponse struct {
	response.CommonBase

	TotalCount int

	ClusterSet []ClusterSet
}

func (c *UKafkaClient) NewListUKafkaInstance() *ListUKafkaInstance {
	req := &ListUKafkaInstance{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

func (c *UKafkaClient) ListUKafka(req *ListUKafkaInstance) (*ListUKafkaInstanceResponse, error) {
	var err error
	var res ListUKafkaInstanceResponse

	err = c.Client.InvokeAction("ListUKafkaInstance", req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
