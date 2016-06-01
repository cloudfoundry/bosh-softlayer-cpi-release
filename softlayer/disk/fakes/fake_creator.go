package fakes

import (
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
)

type FakeCreator struct {
	CreateSize int
	CreateDisk bslcdisk.Disk
	CreateErr  error
}

func (c *FakeCreator) Create(size int, diskCloudProperties bslcdisk.DiskCloudProperties, virtualGuestId int) (bslcdisk.Disk, error) {
	c.CreateSize = size
	return c.CreateDisk, c.CreateErr
}
