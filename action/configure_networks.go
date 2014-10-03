package action

import (
	"fmt"
)

type ConfigureNetworks struct{}

func NewConfigureNetworks() ConfigureNetworks {
	return ConfigureNetworks{}
}

func (a ConfigureNetworks) Run(vmCID VMCID, networks Networks) (interface{}, error) {
	//DEBUG
	fmt.Println("ConfigureNetworks.Run")
	fmt.Printf("----> vmCID: %#v\n", vmCID)
	fmt.Printf("----> networks: %#v\n", networks)
	fmt.Println()
	//DEBUG

	return nil, nil
}
