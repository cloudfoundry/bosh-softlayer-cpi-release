package fakes

import (
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

type FakeFinder struct {
	FindID       string
	FindStemcell bslcstem.Stemcell
	FindFound    bool
	FindErr      error
}

func (f *FakeFinder) Find(id string) (bslcstem.Stemcell, bool, error) {
	f.FindID = id
	return f.FindStemcell, f.FindFound, f.FindErr
}
