package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	disk "bosh-softlayer-cpi/softlayer/disk_service"
)

type SetDiskMetadata struct {
	diskService disk.Service
}

func NewSetDiskMetadata(
	diskService disk.Service,
) SetDiskMetadata {
	return SetDiskMetadata{
		diskService: diskService,
	}
}

func (sdm SetDiskMetadata) Run(DiskCID DiskCID, diskMetadata DiskMetadata) (interface{}, error) {
	if err := sdm.diskService.SetMetadata(DiskCID.Int(), disk.Metadata(diskMetadata)); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Setting metadata for vm '%s'", DiskCID)
	}

	return nil, nil
}
