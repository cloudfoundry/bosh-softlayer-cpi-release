package disk

type Creator interface {
	Create(size int) (Disk, error)
}

type Finder interface {
	Find(id string) (Disk, bool, error)
}

type Disk interface {
	ID() string
	Path() string

	Delete() error
}
