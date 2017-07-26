package action

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type StemcellCID int
type VMCID int
type DiskCID int
type SnapshotCID int

func (stemcellCID StemcellCID) Int() int {
	return int(stemcellCID)
}

func (stemcellCID StemcellCID) String() string {
	return strconv.Itoa(int(stemcellCID))
}

func (stemcellCID *StemcellCID) UnmarshalJSON(data []byte) error {
	if stemcellCID == nil {
		return errors.New("StemcellCID: UnmarshalJSON on nil pointer")
	}

	dataString := strings.Trim(string(data), "\"")
	intValue, err := strconv.Atoi(dataString)
	if err != nil {
		return err
	}
	*stemcellCID = StemcellCID(intValue)

	return nil
}

func (stemcellCID StemcellCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(stemcellCID))
}

func (vmCID VMCID) String() string {
	return strconv.Itoa(int(vmCID))
}

func (vmCID VMCID) Int() int {
	return int(vmCID)
}

func (vmCID *VMCID) UnmarshalJSON(data []byte) error {
	if vmCID == nil {
		return errors.New("VMCID: UnmarshalJSON on nil pointer")
	}

	dataString := strings.Trim(string(data), "\"")
	intValue, err := strconv.Atoi(dataString)
	if err != nil {
		return err
	}

	*vmCID = VMCID(intValue)

	return nil
}

func (vmCID VMCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(vmCID))
}

func (diskCID *DiskCID) UnmarshalJSON(data []byte) error {
	if diskCID == nil {
		return errors.New("DiskCID: UnmarshalJSON on nil pointer")
	}

	dataString := strings.Trim(string(data), "\"")
	intValue, err := strconv.Atoi(dataString)
	if err != nil {
		return err
	}

	*diskCID = DiskCID(intValue)

	return nil
}

func (diskCID DiskCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(diskCID))
}

func (diskCID DiskCID) String() string {
	return strconv.Itoa(int(diskCID))
}

func (diskCID DiskCID) Int() int {
	return int(diskCID)
}

func (snapshotCID *SnapshotCID) UnmarshalJSON(data []byte) error {
	if snapshotCID == nil {
		return errors.New("SnapshotCID: UnmarshalJSON on nil pointer")
	}

	dataString := strings.Trim(string(data), "\"")
	intValue, err := strconv.Atoi(dataString)
	if err != nil {
		return err
	}

	*snapshotCID = SnapshotCID(intValue)

	return nil
}

func (snapshotCID SnapshotCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(snapshotCID))
}

func (snapshotCID SnapshotCID) String() string {
	return strconv.Itoa(int(snapshotCID))
}

func (snapshotCID SnapshotCID) Int() int {
	return int(snapshotCID)
}
