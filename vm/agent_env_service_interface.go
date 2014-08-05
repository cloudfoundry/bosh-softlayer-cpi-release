package vm

type AgentEnvService interface {
	// Fetch will return an error if Update was not called beforehand
	Fetch() (AgentEnv, error)
	Update(AgentEnv) error
}
