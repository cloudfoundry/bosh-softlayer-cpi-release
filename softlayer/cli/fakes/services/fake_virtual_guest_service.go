package fake_softlayer_service

import (
	"encoding/json"
	test_helpers "github.com/maximilien/bosh-softlayer-cpi/common"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	"os"
	"path/filepath"
)

type fake_Virtual_Guest_Service struct {
	client softlayer.Client
}

func NewFake_Virtual_Guest_Service(client softlayer.Client) *fake_Virtual_Guest_Service {
	return &fake_Virtual_Guest_Service{
		client: client,
	}
}

func (fvgs *fake_Virtual_Guest_Service) GetName() string {
	return "Fake_Virtual_Guest"
}

func (fvgs *fake_Virtual_Guest_Service) CreateObject(template datatypes.SoftLayer_Virtual_Guest_Template) (datatypes.SoftLayer_Virtual_Guest, error) {
	fake_Virtual_Guest := &datatypes.SoftLayer_Virtual_Guest{}
	workingDir, err := os.Getwd()
	data, err := test_helpers.ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", "SoftLayer_Virtual_Guest_Service_createObject.json")
	json.Unmarshal(data, fake_Virtual_Guest)
	return *fake_Virtual_Guest, err
}

func (fvgs *fake_Virtual_Guest_Service) GetObject(instanceId int) (datatypes.SoftLayer_Virtual_Guest, error) {
	return datatypes.SoftLayer_Virtual_Guest{}, nil
}

func (fvgs *fake_Virtual_Guest_Service) EditObject(instanceId int, template datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	return true, nil
}

func (fvgs *fake_Virtual_Guest_Service) DeleteObject(instanceId int) (bool, error) {
	return true, nil
}

func (fvgs *fake_Virtual_Guest_Service) GetPowerState(instanceId int) (datatypes.SoftLayer_Virtual_Guest_Power_State, error) {
	return datatypes.SoftLayer_Virtual_Guest_Power_State{}, nil
}
func (fvgs *fake_Virtual_Guest_Service) GetSshKeys(instanceId int) ([]datatypes.SoftLayer_Security_Ssh_Key, error) {
	return []datatypes.SoftLayer_Security_Ssh_Key{}, nil
}

func (fvgs *fake_Virtual_Guest_Service) GetActiveTransaction(instanceId int) (datatypes.SoftLayer_Provisioning_Version1_Transaction, error) {
	return datatypes.SoftLayer_Provisioning_Version1_Transaction{}, nil
}
func (fvgs *fake_Virtual_Guest_Service) GetActiveTransactions(instanceId int) ([]datatypes.SoftLayer_Provisioning_Version1_Transaction, error) {
	return []datatypes.SoftLayer_Provisioning_Version1_Transaction{}, nil
}
func (fvgs *fake_Virtual_Guest_Service) RebootSoft(instanceId int) (bool, error) {
	return true, nil
}
func (fvgs *fake_Virtual_Guest_Service) RebootHard(instanceId int) (bool, error) {
	return true, nil
}
func (fvgs *fake_Virtual_Guest_Service) SetMetadata(instanceId int, metadata string) (bool, error) {
	return true, nil
}
func (fvgs *fake_Virtual_Guest_Service) ConfigureMetadataDisk(instanceId int) (datatypes.SoftLayer_Provisioning_Version1_Transaction, error) {
	return datatypes.SoftLayer_Provisioning_Version1_Transaction{}, nil
}

func (fvgs *fake_Virtual_Guest_Service) AttachIscsiVolume(instanceId int, volumeId int) (string, error) {
	return "", nil
}

func (fvgs *fake_Virtual_Guest_Service) DetachIscsiVolume(instanceId int, volumeId int) error {
	return nil
}
