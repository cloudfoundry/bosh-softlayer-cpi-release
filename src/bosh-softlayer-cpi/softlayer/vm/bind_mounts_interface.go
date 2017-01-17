package vm

type GuestBindMounts interface {
	MakeEphemeral() string
	MakePersistent() string
	MountPersistent(diskID string) string
}

type HostBindMounts interface {
	MakeEphemeral(id string) (string, error)
	DeleteEphemeral(id string) error

	MakePersistent(id string) (string, error)
	DeletePersistent(id string) error

	MountPersistent(id, diskID, diskPath string) error
	UnmountPersistent(id, diskID string) error
}
