package cpi_fakes

import (
	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
)

type FakeClient struct {
	FakeContainer  Container
	FakeContainers []Container
	FakeError      error

	Connection FakeConnection
}

func New() *FakeClient {
	return &FakeClient{}
}

func (fc *FakeClient) Create(containerSpec ContainerSpec) (Container, error) {
	return fc.FakeContainer, fc.FakeError
}

func (fc *FakeClient) Destroy(vmId string) error {
	return fc.FakeError
}

func (fc *FakeClient) Containers() ([]Container, error) {
	return fc.FakeContainers, fc.FakeError
}
