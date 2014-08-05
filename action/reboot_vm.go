package action

type RebootVM struct{}

func NewRebootVM() RebootVM {
	return RebootVM{}
}

func (a RebootVM) Run(vmCID VMCID) (interface{}, error) {
	return nil, nil
}
