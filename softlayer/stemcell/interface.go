package stemcell

type Finder interface {
	Find(uuid string) (Stemcell, bool, error)
	FindById(id int) (Stemcell, bool, error)
}

type Stemcell interface {
	ID() int
	Uuid() string
	Kind() string

	Delete() error
}
