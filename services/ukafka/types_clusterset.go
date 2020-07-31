package ukafka

type ClusterSet struct {
	Zone string

	ClusterInstanceId string

	ClusterInstanceName string

	Framework string

	FrameworkVersion string

	Remark string

	CreateTime int

	RunningTime int
	ExpireTime int

	AutoRenew string

	ChargeType string

	UHostCount int

	RedundantCount int

	State string

	Tag string

	NewMessage string
	// 所在的VPC的ID
	VPCId string

	// 为 InnerMode 时，ULB 所属的子网ID，默认为空
	SubnetId string

	// 所属的业务组ID
	BusinessId string
}
