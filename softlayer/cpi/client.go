package cpi

type Client interface {
	Create(containerSpec ContainerSpec) (Container, error)
	Destroy(vmId string) error

	Containers() ([]Container, error)
}
