package vm

import (
	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
)

type AgentEnvServiceFactory interface {
	New(bslcpi.Container) AgentEnvService
}
