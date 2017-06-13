package client

import (
	slpoolclient "bosh-softlayer-cpi/softlayer/vps_service/client"
)

func NewSoftLayerVmPoolClient() slpoolclient.SoftLayerVMPool {
	return *slpoolclient.NewHTTPClient(nil)
}
