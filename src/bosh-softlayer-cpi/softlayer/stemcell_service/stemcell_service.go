package stemcell

type Service interface {
	Find(id int) (string, bool, error)
}
