package stemcell

//go:generate counterfeiter -o fakes/fake_Stemcell_Service.go . StemcellService
type Service interface {
	Find(id int) (string, bool, error)
}
