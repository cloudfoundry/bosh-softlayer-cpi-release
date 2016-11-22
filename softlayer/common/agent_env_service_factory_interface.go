package common

type AgentEnvServiceFactory interface {
	New(VM, SoftlayerFileService) AgentEnvService
}
