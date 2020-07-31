package usocketmon

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UMonClient is the client of UMon
type USocketMonClient struct {
	*ucloud.Client
}

// NewClient will return a instance of USocketMonClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *USocketMonClient {
	client := ucloud.NewClient(config, credential)
	return &USocketMonClient{
		client,
	}
}
