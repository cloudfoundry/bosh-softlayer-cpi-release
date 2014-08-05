package stemcell

type Importer interface {
	ImportFromPath(imagePath string) (Stemcell, error)
}

type Finder interface {
	Find(id string) (Stemcell, bool, error)
}

type Stemcell interface {
	ID() string
	DirPath() string

	Delete() error
}
