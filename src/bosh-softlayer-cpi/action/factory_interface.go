package action

//go:generate counterfeiter -o fakes/fake_factory.go . Factory
type Factory interface {
	Create(method string) (Action, error)
}
