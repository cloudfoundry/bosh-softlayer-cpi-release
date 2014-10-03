package action

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type HasVM struct {
	vmFinder bslcvm.Finder
}

func NewHasVM(vmFinder bslcvm.Finder) HasVM {
	return HasVM{vmFinder: vmFinder}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	//DEBUG
	fmt.Println("HasVM.Run")
	fmt.Printf("----> vmCID: %#v\n", vmCID)
	fmt.Println()
	//DEBUG

	_, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return false, bosherr.WrapError(err, "Finding VM '%s'", vmCID)
	}

	return found, nil
}
