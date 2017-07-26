package stemcell

import (
	bosl "bosh-softlayer-cpi/softlayer/client"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SoftlayerStemcellService struct {
	softlayerClient bosl.Client
	logger          boshlog.Logger
}

func NewSoftlayerStemcellService(
	softlayerClient bosl.Client,
	logger boshlog.Logger,
) SoftlayerStemcellService {
	return SoftlayerStemcellService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
