package action

import (
	"fmt"
)

type RebootVM struct{}

func NewRebootVM() RebootVM {
	return RebootVM{}
}

func (a RebootVM) Run(vmCID VMCID) (interface{}, error) {
	//DEBUG
	fmt.Println("RebootVM.Run")
	fmt.Printf("----> vmCID: %#v\n", vmCID)
	fmt.Println()
	//DEBUG

	return nil, nil
}
