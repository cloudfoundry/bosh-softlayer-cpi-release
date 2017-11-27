package stemcell

import (
	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

type SoftlayerStemcellService struct {
	softlayerClient bosl.Client
	logger          logger.Logger
}

func NewSoftlayerStemcellService(
	softlayerClient bosl.Client,
	logger logger.Logger,
) SoftlayerStemcellService {
	return SoftlayerStemcellService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
