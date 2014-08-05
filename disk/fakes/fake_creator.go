package fakes

import (
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/disk"
)

type FakeCreator struct {
	CreateSize int
	CreateDisk bslcdisk.Disk
	CreateErr  error
}

func (c *FakeCreator) Create(size int) (bslcdisk.Disk, error) {
	c.CreateSize = size
	return c.CreateDisk, c.CreateErr
}
