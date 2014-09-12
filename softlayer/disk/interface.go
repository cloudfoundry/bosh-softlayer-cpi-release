package disk

type Creator interface {
	Create(size int) (Disk, error)
}

type Finder interface {
	Find(id int) (Disk, bool, error)
}

type Disk interface {
	ID() int
	Path() string

	Delete() error
}
