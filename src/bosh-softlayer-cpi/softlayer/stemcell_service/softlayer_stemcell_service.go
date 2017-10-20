package stemcell

import (
	bosl "bosh-softlayer-cpi/softlayer/client"
	"bosh-softlayer-cpi/logger"
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
