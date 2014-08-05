package fakes

type FakeDisk struct {
	id   string
	path string

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeDisk(id string) *FakeDisk {
	return &FakeDisk{id: id}
}

func NewFakeDiskWithPath(id, path string) *FakeDisk {
	return &FakeDisk{id: id, path: path}
}

func (s FakeDisk) ID() string { return s.id }

func (s FakeDisk) Path() string { return s.path }

func (s *FakeDisk) Delete() error {
	s.DeleteCalled = true
	return s.DeleteErr
}
