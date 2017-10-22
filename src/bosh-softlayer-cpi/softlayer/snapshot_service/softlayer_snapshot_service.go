package snapshot

import (
	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

const softlayerSnapshotServiceLogTag = "SoftLayerDiskCreator"
const softlayerSnapshotDescription = "Snapshot_managed_by_BOSH"

type SoftlayerSnapshotService struct {
	softlayerClient bosl.Client
	logger          logger.Logger
}

func NewSoftlayerSnapshotService(
	softlayerClient bosl.Client,
	logger logger.Logger,
) SoftlayerSnapshotService {
	return SoftlayerSnapshotService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}
