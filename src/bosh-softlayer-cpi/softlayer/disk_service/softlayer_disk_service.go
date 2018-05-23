package disk

import (
	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

const softlayerDiskServiceLogTag = "SoftLayerDiskCreator"

type SoftlayerDiskService struct {
	softlayerClient bosl.Client
	logger          logger.Logger
}

func NewSoftlayerDiskService(
	softlayerClient bosl.Client,
	logger logger.Logger,
) SoftlayerDiskService {
	return SoftlayerDiskService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
