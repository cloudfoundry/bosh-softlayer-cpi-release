package disk

type DiskCloudProperties struct {
	ConsistentPerformanceIscsi bool `json:"consistent_performance_iscsi,omitempty"`
}

type Creator interface {
	Create(size int, cloudProp DiskCloudProperties, virtualGuestId int) (Disk, error)
}

type Finder interface {
	Find(id int) (Disk, bool, error)
}

type Disk interface {
	ID() int

	Delete() error
}
