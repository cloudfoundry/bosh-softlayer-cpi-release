package fakes

import (
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/vm"
)

type FakeCreator struct {
	CreateAgentID     string
	CreateStemcell    bslcstem.Stemcell
	CreateNetworks    bslcvm.Networks
	CreateEnvironment bslcvm.Environment
	CreateVM          bslcvm.VM
	CreateErr         error
}

func (c *FakeCreator) Create(agentID string, stemcell bslcstem.Stemcell, networks bslcvm.Networks, env bslcvm.Environment) (bslcvm.VM, error) {
	c.CreateAgentID = agentID
	c.CreateStemcell = stemcell
	c.CreateNetworks = networks
	c.CreateEnvironment = env
	return c.CreateVM, c.CreateErr
}
