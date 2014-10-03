package action

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type StemcellCID string
type VMCID int
type DiskCID int

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
	return json.Marshal(struct {
		VMCID int `json:"vmCID"`
	}{
		VMCID: int(vmCID),
	})
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
	return json.Marshal(struct {
		DiskCID int `json:"diskCID"`
	}{
		DiskCID: int(diskCID),
	})
}
