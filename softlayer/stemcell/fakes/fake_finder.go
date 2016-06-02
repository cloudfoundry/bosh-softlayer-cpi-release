package fakes

import (
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
)

type FakeFinder struct {
	FindID       int
	FindUuid     string
	FindStemcell bslcstem.Stemcell
	FindFound    bool
	FindErr      error
}

func (f *FakeFinder) Find(uuid string) (bslcstem.Stemcell, bool, error) {
	f.FindUuid = uuid
	return f.FindStemcell, f.FindFound, f.FindErr
}

func (f *FakeFinder) FindById(id int) (bslcstem.Stemcell, bool, error) {
	f.FindID = id
	return f.FindStemcell, f.FindFound, f.FindErr
}
