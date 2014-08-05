package fakes

type FakeStemcell struct {
	id      string
	dirPath string

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeStemcell(id string) *FakeStemcell {
	return &FakeStemcell{id: id}
}

func NewFakeStemcellWithPath(id, dirPath string) *FakeStemcell {
	return &FakeStemcell{id: id, dirPath: dirPath}
}

func (s FakeStemcell) ID() string { return s.id }

func (s FakeStemcell) DirPath() string { return s.dirPath }

func (s *FakeStemcell) Delete() error {
	s.DeleteCalled = true
	return s.DeleteErr
}
