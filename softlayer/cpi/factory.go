package cpi

type SoftLayerConnection struct {
	Network string
	Address string
}

func NewConnection(network, address string) SoftLayerConnection {
	return SoftLayerConnection{
		Network: network,
		Address: address,
	}
}

func NewClient(slConnection SoftLayerConnection) Client {
	return nil
}
