package fake_softlayer_service

import (
	"encoding/json"
	"fmt"
	test_helpers "github.com/maximilien/bosh-softlayer-cpi/common"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	"os"
	"path/filepath"
)

type fake_Network_Storage_Service struct {
	client softlayer.Client
}

func NewFake_Network_Storage_Service(client softlayer.Client) *fake_Network_Storage_Service {
	return &fake_Network_Storage_Service{
		client: client,
	}
}

func (fvgs *fake_Network_Storage_Service) GetName() string {
	return "Fake_Network_Storage"
}

func (fnss *fake_Network_Storage_Service) CreateIscsiVolume(size int, location string) (datatypes.SoftLayer_Network_Storage, error) {
	fmt.Println("_____")
	fake_Network_Storage := &datatypes.SoftLayer_Network_Storage{}
	workingDir, err := os.Getwd()
	data, err := test_helpers.ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", "SoftLayer_Network_Storage_Service_createIscsiVolume.json")
	json.Unmarshal(data, fake_Network_Storage)
	return *fake_Network_Storage, err
}

func (fnss *fake_Network_Storage_Service) DeleteIscsiVolume(volumeId int, immediateCancellationFlag bool) error {
	return nil
}

func (fnss *fake_Network_Storage_Service) GetIscsiVolume(volumeId int) (datatypes.SoftLayer_Network_Storage, error) {
	return datatypes.SoftLayer_Network_Storage{}, nil
}
