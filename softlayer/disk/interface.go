package disk

type DiskCloudProperties struct {
	Iops             int  `json:"iops,omitempty"`
	UseHourlyPricing bool `json:"useHourlyPricing,omitempty"`
}

//go:generate counterfeiter -o fakes/fake_disk_creator.go . Creator
type Creator interface {
	Create(size int, cloudProp DiskCloudProperties, datacenter_id int) (Disk, error)
}

//go:generate counterfeiter -o fakes/fake_disk_finder.go . Finder
type Finder interface {
	Find(id int) (Disk, bool, error)
}

//go:generate counterfeiter -o fakes/fake_disk.go . Disk
type Disk interface {
	ID() int

	Delete() error
}
