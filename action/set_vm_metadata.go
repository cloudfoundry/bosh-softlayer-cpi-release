package action

import "fmt"

type SetVMMetadata struct{}

type Bar struct {
	Test int `json:"test"`
}

type VMMetadata struct {
	Foo string `json:"foo,omitempty"`
	Bar Bar    `json:"bar,omitempty"`
}

func NewSetVMMetadata() SetVMMetadata {
	return SetVMMetadata{}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata VMMetadata) (interface{}, error) {
	//DEBUG
	fmt.Println("SetVMMetadata.Run")
	fmt.Printf("----> vmCID: %#v\n", vmCID)
	fmt.Printf("----> metadata: %#v\n", metadata)
	fmt.Println()
	//DEBUG

	return nil, nil
}
