package disk

type DiskCloudProperties struct {
	Iops             int  `json:"iops,omitempty"`
	UseHourlyPricing bool `json:"useHourlyPricing,omitempty"`
}

type Creator interface {
	Create(size int, cloudProp DiskCloudProperties, datacenter_id int) (Disk, error)
}

type Finder interface {
	Find(id int) (Disk, bool, error)
}

type Disk interface {
	ID() int

	Delete() error
}
