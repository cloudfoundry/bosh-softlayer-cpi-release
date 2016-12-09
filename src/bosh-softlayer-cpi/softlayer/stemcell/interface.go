package stemcell

//go:generate counterfeiter -o fakes/fake_stemcell_finder.go . StemcellFinder
type StemcellFinder interface {
	FindById(id int) (Stemcell, error)
}

//go:generate counterfeiter -o fakes/fake_stemcell.go . Stemcell
type Stemcell interface {
	ID() int
	Uuid() string

	Delete() error
}
