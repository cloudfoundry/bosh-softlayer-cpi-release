package fakes

const FakeStemcellKind = "FakeStemcellKind"

type FakeStemcell struct {
	id   int
	uuid string
	kind string

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeStemcell(id int, uuid string, kind string) *FakeStemcell {
	return &FakeStemcell{
		id:   id,
		uuid: uuid,
		kind: kind,
	}
}

func (s FakeStemcell) ID() int { return s.id }

func (s FakeStemcell) Uuid() string { return s.uuid }

func (s FakeStemcell) Kind() string { return s.kind }

func (s *FakeStemcell) Delete() error {
	s.DeleteCalled = true
	return s.DeleteErr
}
