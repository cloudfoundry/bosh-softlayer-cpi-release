package fakes

import (
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type FakeCreator struct {
	CreateAgentID           string
	CreateStemcell          bslcstem.Stemcell
	CreateNetworks          bslcvm.Networks
	CreateVMCloudProperties bslcvm.VMCloudProperties
	CreateEnvironment       bslcvm.Environment
	CreateVM                bslcvm.VM
	CreateErr               error
}

func (c *FakeCreator) Create(agentID string, stemcell bslcstem.Stemcell, vmCloudProperties bslcvm.VMCloudProperties, networks bslcvm.Networks, env bslcvm.Environment) (bslcvm.VM, error) {
	c.CreateAgentID = agentID
	c.CreateStemcell = stemcell
	c.CreateVMCloudProperties = vmCloudProperties
	c.CreateNetworks = networks
	c.CreateEnvironment = env
	return c.CreateVM, c.CreateErr
}
