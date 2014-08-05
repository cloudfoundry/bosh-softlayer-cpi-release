package action

type SetVMMetadata struct{}

type VMMetadata struct{}

func NewSetVMMetadata() SetVMMetadata {
	return SetVMMetadata{}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata VMMetadata) (interface{}, error) {
	// todo can properties be set on the container
	return nil, nil
}
