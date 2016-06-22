package vm

type AgentEnvServiceFactory interface {
	New(VM, SoftlayerFileService) AgentEnvService
}
