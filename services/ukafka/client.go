package ukafka

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UMonClient is the client of UMon
type UKafkaClient struct {
	*ucloud.Client
}

// NewClient will return a instance of USocketMonClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UKafkaClient {
	client := ucloud.NewClient(config, credential)
	return &UKafkaClient{
		client,
	}
}
