package disk

import (
	bosl "bosh-softlayer-cpi/softlayer/client"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const softlayerDiskServiceLogTag = "SoftLayerDiskCreator"

type SoftlayerDiskService struct {
	softlayerClient bosl.Client
	logger          boshlog.Logger
}

func NewSoftlayerDiskService(
	softlayerClient bosl.Client,
	logger boshlog.Logger,
) SoftlayerDiskService {
	return SoftlayerDiskService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
