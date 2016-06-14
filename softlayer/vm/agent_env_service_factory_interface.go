package vm

type AgentEnvServiceFactory interface {
	New(SoftlayerFileService) AgentEnvService
}
