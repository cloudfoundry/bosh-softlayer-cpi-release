package action

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"reflect"
)

type StemcellCID int
type VMCID int
type DiskCID int

func (stemcellCID StemcellCID) String() string {
	return strconv.Itoa(int(stemcellCID))
}

func (stemcellCID *StemcellCID) UnmarshalJSON(data []byte) error {
	if stemcellCID == nil {
		return errors.New("StemcellCID: UnmarshalJSON on nil pointer")
	}

	dataString := strings.Trim(string(data), "\"")
	if isNumber(dataString) {
		intValue, err := strconv.Atoi(dataString)
		if err != nil {
			return err
		}
		*stemcellCID = StemcellCID(intValue)
	}

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

func isNumber(a interface{}) bool {
	if a == nil {
		return false
	}
	kind := reflect.TypeOf(a).Kind()
	return reflect.Int <= kind && kind <= reflect.Float64
}

