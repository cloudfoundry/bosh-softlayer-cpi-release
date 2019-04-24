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
	_, err := dd.diskService.Find(diskCID.Int())
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, nil
		}
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s' to vm '%s", diskCID)
	}
	return nil, dd.diskService.Delete(diskCID.Int())
}
