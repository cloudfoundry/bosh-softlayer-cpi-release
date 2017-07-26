package action

import (
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
	return nil, dd.diskService.Delete(diskCID.Int())
}
