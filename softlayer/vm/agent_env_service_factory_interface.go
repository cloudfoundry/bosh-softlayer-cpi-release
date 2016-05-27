package vm

type AgentEnvServiceFactory interface {
	New(SoftlayerFileService, string) AgentEnvService
}
