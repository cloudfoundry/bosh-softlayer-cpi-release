package fakes

import (
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
)

type FakeFinder struct {
	FindID       int
	FindUuid     string
	FindStemcell bslcstem.Stemcell
	FindErr      error
}

func (f *FakeFinder) FindById(id int) (bslcstem.Stemcell, error) {
	f.FindID = id
	return f.FindStemcell, f.FindErr
}
