package fakes

import (
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/disk"
)

type FakeFinder struct {
	FindID    string
	FindDisk  bslcdisk.Disk
	FindFound bool
	FindErr   error
}

func (f *FakeFinder) Find(id string) (bslcdisk.Disk, bool, error) {
	f.FindID = id
	return f.FindDisk, f.FindFound, f.FindErr
}
