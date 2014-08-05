package vm

import (
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
)

type AgentEnvServiceFactory interface {
	New(wrdn.Container) AgentEnvService
}
