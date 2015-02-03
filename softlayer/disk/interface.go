package disk

type Creator interface {
	Create(size int, virtualGuestId int) (Disk, error)
}

type Finder interface {
	Find(id int) (Disk, bool, error)
}

type Disk interface {
	ID() int

	Delete() error
}
