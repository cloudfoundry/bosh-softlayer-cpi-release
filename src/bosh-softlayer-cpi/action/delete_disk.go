package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/disk_service"
)

type DeleteDisk struct {
	diskService disk.Service
}

func NewDeleteDisk(
	diskService disk.Service,
) DeleteDisk {
	return DeleteDisk{
		diskService: diskService,
	}
}

func (dd DeleteDisk) Run(diskCID DiskCID) (interface{}, error) {
	if err := dd.diskService.Delete(diskCID.Int()); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Deleting disk '%s'", diskCID)
	}

	return nil, nil
}
