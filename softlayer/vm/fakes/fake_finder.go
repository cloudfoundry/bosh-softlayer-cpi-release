package fakes

import (
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type FakeFinder struct {
	FindID    string
	FindVM    bslcvm.VM
	FindFound bool
	FindErr   error
}

func (f *FakeFinder) Find(id string) (bslcvm.VM, bool, error) {
	f.FindID = id
	return f.FindVM, f.FindFound, f.FindErr
}
