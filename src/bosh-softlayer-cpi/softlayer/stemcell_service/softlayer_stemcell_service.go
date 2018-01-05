package stemcell

import (
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

const softlayerImageNamePrefix = "stemcell"
const softlayerStemcellServiceLogTag = "SoftlayerStemcellService"

type SoftlayerStemcellService struct {
	softlayerClient bosl.Client
	uuidGen         boshuuid.Generator
	logger          logger.Logger
}

func NewSoftlayerStemcellService(
	softlayerClient bosl.Client,
	uuidGen boshuuid.Generator,
	logger logger.Logger,
) SoftlayerStemcellService {
	return SoftlayerStemcellService{
		softlayerClient: softlayerClient,
		uuidGen:         uuidGen,
		logger:          logger,
	}
}
