package instance

import (
	bosl "bosh-softlayer-cpi/softlayer/client"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const rootUser = "root"
const softlayerVirtualGuestServiceLogTag = "SoftlayerVirtualGuestService"

type SoftlayerVirtualGuestService struct {
	softlayerClient bosl.Client
	uuidGen         boshuuid.Generator
	logger          boshlog.Logger
}

func NewSoftLayerVirtualGuestService(
	softlayerClient bosl.Client,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
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
