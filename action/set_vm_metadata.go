package action

import (
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type SetVMMetadata struct {
	vmFinder bslcvm.Finder
}

func NewSetVMMetadata(vmFinder bslcvm.Finder) SetVMMetadata {
	return SetVMMetadata{vmFinder: vmFinder}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata bslcvm.VMMetadata) (interface{}, error) {
	return nil, nil
}
