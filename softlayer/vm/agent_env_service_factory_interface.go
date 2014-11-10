package vm

type AgentEnvServiceFactory interface {
	New(vmId int) AgentEnvService
}
