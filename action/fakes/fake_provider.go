package fakes

import (
	bosherror "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	fakevm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/fakes"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type fakeProvider struct {
	fakeCreators map[string]bslcvm.VMCreator
}

func NewFakeProvider() CreatorProvider {

	fakeSoftlayerCreator := &fakevm.FakeCreator{}
	fakeSoftlayerCreator.CreateVM = fakevm.NewFakeVM(1234)

	fakeBaremetalCreator := &fakevm.FakeCreator{}
	fakeBaremetalCreator.CreateVM = fakevm.NewFakeVM(1234)

	return &fakeProvider{
		fakeCreators: map[string]bslcvm.VMCreator{
			"virtualguest": fakeSoftlayerCreator,
			"baremetal":    fakeBaremetalCreator,
		},
	}
}

func (p *fakeProvider) Get(name string) (bslcvm.VMCreator, error) {
	creator, found := p.fakeCreators[name]
	if !found {
		return nil, bosherror.Errorf("Creator %s could not be found", name)
	}
	return creator, nil
}
