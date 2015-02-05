package fakes

type FakeDisk struct {
	id   int
	path string

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeDisk(id int) *FakeDisk {
	return &FakeDisk{id: id}
}

func (s FakeDisk) ID() int { return s.id }

func (s FakeDisk) Path() string { return s.path }

func (s *FakeDisk) Delete() error {
	s.DeleteCalled = true
	return s.DeleteErr
}
