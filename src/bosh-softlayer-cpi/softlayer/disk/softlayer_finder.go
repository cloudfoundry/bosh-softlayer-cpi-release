package disk

import (
	"strings"

	bsl "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SoftLayerFinder struct {
	softLayerClient bsl.Client
	logger          boshlog.Logger
}

func NewSoftLayerDiskFinder(client bsl.Client, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{softLayerClient: client, logger: logger}
}

func (sf SoftLayerFinder) Find(id int) (Disk, error) {
	volume, err := sf.softLayerClient.GetBlockVolumeDetails(id, bsl.VOLUME_DEFAULT_MASK)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return nil, bosherr.WrapErrorf(err, "Failed to find iSCSI volume with id: %d", id)
		}
	}

	if volume.Id == nil {
		return nil, bosherr.WrapErrorf(err, "Finding volume with id: `%d` failed", id)
	}

	return NewSoftLayerDisk(id, sf.softLayerClient, sf.logger), nil
}
