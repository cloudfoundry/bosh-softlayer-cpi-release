package stemcell

//go:generate counterfeiter -o fakes/fake_Stemcell_Service.go . Service
type Service interface {
	Find(id int) (string, error)
	CreateFromTarball(imagePath string, datacenter string, osCode string) (int, error)
}
