package instance

import (
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

const rootUser = "root"
const softlayerVirtualGuestServiceLogTag = "SoftlayerVirtualGuestService"

type SoftlayerVirtualGuestService struct {
	softlayerClient bosl.Client
	uuidGen         boshuuid.Generator
	logger          logger.Logger
}

func NewSoftLayerVirtualGuestService(
	softlayerClient bosl.Client,
	uuidGen boshuuid.Generator,
	logger logger.Logger,
) SoftlayerVirtualGuestService {
	return SoftlayerVirtualGuestService{
		softlayerClient: softlayerClient,
		uuidGen:         uuidGen,
		logger:          logger,
	}
}

type Mount struct {
	PartitionPath string
	MountPoint    string
}
