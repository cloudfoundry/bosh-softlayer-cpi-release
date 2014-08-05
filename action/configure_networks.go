package action

type ConfigureNetworks struct{}

func NewConfigureNetworks() ConfigureNetworks {
	return ConfigureNetworks{}
}

func (a ConfigureNetworks) Run(vmCID VMCID, networks Networks) (interface{}, error) {
	return nil, nil
}
