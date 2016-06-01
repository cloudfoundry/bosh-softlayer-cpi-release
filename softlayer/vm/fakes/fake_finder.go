package fakes

import (
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type FakeFinder struct {
	FindID    int
	FindVM    bslcvm.VM
	FindFound bool
	FindErr   error
}

func (f *FakeFinder) Find(id int) (bslcvm.VM, bool, error) {
	f.FindID = id
	return f.FindVM, f.FindFound, f.FindErr
}
