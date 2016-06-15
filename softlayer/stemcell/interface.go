package stemcell

type Finder interface {
	FindById(id int) (Stemcell, error)
}

type Stemcell interface {
	ID() int
	Uuid() string

	Delete() error
}
