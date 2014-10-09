package stemcell

type Finder interface {
	Find(id string) (Stemcell, bool, error)
}

type Stemcell interface {
	ID() string

	Delete() error
}
