package vm

type AgentEnvServiceFactory interface {
	New() AgentEnvService
}
