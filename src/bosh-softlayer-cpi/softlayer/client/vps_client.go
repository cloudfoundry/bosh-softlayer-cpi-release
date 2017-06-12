package client

import (
	slpoolclient "bosh-softlayer-cpi/softlayer/pool/client"
)

func NewSoftLayerVmPoolClient() slpoolclient.SoftLayerVMPool {
	return *slpoolclient.NewHTTPClient(nil)
}
