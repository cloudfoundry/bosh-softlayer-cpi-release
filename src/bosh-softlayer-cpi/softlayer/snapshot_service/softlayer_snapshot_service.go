package snapshot

import (
	bosl "bosh-softlayer-cpi/softlayer/client"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const softlayerSnapshotServiceLogTag = "SoftLayerDiskCreator"
const softlayerSnapshotDescription = "Snapshot_managed_by_BOSH"

type SoftlayerSnapshotService struct {
	softlayerClient bosl.Client
	logger          boshlog.Logger
}

func NewSoftlayerSnapshotService(
	softlayerClient bosl.Client,
	logger boshlog.Logger,
) SoftlayerSnapshotService {
	return SoftlayerSnapshotService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
