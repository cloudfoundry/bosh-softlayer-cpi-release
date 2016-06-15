package fakes

type FakeStemcell struct {
	id   int
	uuid string

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeStemcell(id int, uuid string) *FakeStemcell {
	return &FakeStemcell{
		id:   id,
		uuid: uuid,
	}
}

func (s FakeStemcell) ID() int { return s.id }

func (s FakeStemcell) Uuid() string { return s.uuid }

func (s *FakeStemcell) Delete() error {
	s.DeleteCalled = true
	return s.DeleteErr
}
