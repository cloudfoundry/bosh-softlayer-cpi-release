package snapshot

//go:generate counterfeiter -o fakes/fake_Snapshot_Service.go . Service
type Service interface {
	Create(diskID int, description string) (int, error)
	Delete(id int) error
}
