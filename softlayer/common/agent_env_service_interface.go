package common

//go:generate counterfeiter -o fakes/fake_agent_env_service.go . AgentEnvService
type AgentEnvService interface {
	Fetch() (AgentEnv, error)
	Update(AgentEnv) error
}
