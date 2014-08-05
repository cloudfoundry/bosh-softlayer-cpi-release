package fakes

import (
	"path/filepath"
)

type FakeGuestBindMounts struct {
	EphemeralBindMountPath  string
	PersistentBindMountsDir string
}

func (gbm FakeGuestBindMounts) MakeEphemeral() string {
	return gbm.EphemeralBindMountPath
}

func (gbm FakeGuestBindMounts) MakePersistent() string {
	return gbm.PersistentBindMountsDir
}

func (gbm FakeGuestBindMounts) MountPersistent(diskID string) string {
	return filepath.Join(gbm.PersistentBindMountsDir, diskID)
}

type FakeHostBindMounts struct {
	MakeEphemeralID   string
	MakeEphemeralPath string
	MakeEphemeralErr  error

	DeleteEphemeralID  string
	DeleteEphemeralErr error

	MakePersistentID   string
	MakePersistentPath string
	MakePersistentErr  error

	DeletePersistentID  string
	DeletePersistentErr error

	MountPersistentID       string
	MountPersistentDiskID   string
	MountPersistentDiskPath string
	MountPersistentErr      error

	UnmountPersistentID     string
	UnmountPersistentDiskID string
	UnmountPersistentErr    error
}

func (hbm *FakeHostBindMounts) MakeEphemeral(id string) (string, error) {
	hbm.MakeEphemeralID = id
	return hbm.MakeEphemeralPath, hbm.MakeEphemeralErr
}

func (hbm *FakeHostBindMounts) DeleteEphemeral(id string) error {
	hbm.DeleteEphemeralID = id
	return hbm.DeleteEphemeralErr
}

func (hbm *FakeHostBindMounts) MakePersistent(id string) (string, error) {
	hbm.MakePersistentID = id
	return hbm.MakePersistentPath, hbm.MakePersistentErr
}

func (hbm *FakeHostBindMounts) DeletePersistent(id string) error {
	hbm.DeletePersistentID = id
	return hbm.DeletePersistentErr
}

func (hbm *FakeHostBindMounts) MountPersistent(id, diskID, diskPath string) error {
	hbm.MountPersistentID = id
	hbm.MountPersistentDiskID = diskID
	hbm.MountPersistentDiskPath = diskPath
	return hbm.MountPersistentErr
}

func (hbm *FakeHostBindMounts) UnmountPersistent(id, diskID string) error {
	hbm.UnmountPersistentID = id
	hbm.UnmountPersistentDiskID = diskID
	return hbm.UnmountPersistentErr
}
