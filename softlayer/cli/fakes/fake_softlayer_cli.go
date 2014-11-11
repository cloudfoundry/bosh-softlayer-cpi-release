package fake_softlayer_cli

import (
	"bytes"
	"errors"
	"fmt"
	fake_services "github.com/maximilien/bosh-softlayer-cpi/softlayer/cli/fakes/services"
	. "github.com/maximilien/softlayer-go/softlayer"
)

type Fake_SoftLayer_Client struct {
	SoftLayerServices map[string]Service
}

func NewFake_SoftLayer_Client() *Fake_SoftLayer_Client {
	fslc := &Fake_SoftLayer_Client{
		SoftLayerServices: map[string]Service{},
	}

	fslc.initSoftLayerServices()

	return fslc
}

func (fslc *Fake_SoftLayer_Client) GetService(name string) (Service, error) {
	slService, ok := fslc.SoftLayerServices[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("softlayer-go does not support service '%s'", name))
	}

	return slService, nil
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Account_Service() (SoftLayer_Account_Service, error) {
	service, err := fslc.GetService("SoftLayer_Account")
	return service.(SoftLayer_Account_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Virtual_Guest_Service() (SoftLayer_Virtual_Guest_Service, error) {
	service, err := fslc.GetService("SoftLayer_Virtual_Guest")
	return service.(SoftLayer_Virtual_Guest_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Virtual_Disk_Image_Service() (SoftLayer_Virtual_Disk_Image_Service, error) {
	service, err := fslc.GetService("SoftLayer_Virtual_Disk_Image")
	return service.(SoftLayer_Virtual_Disk_Image_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Security_Ssh_Key_Service() (SoftLayer_Security_Ssh_Key_Service, error) {
	service, err := fslc.GetService("SoftLayer_Security_Ssh_Key")
	return service.(SoftLayer_Security_Ssh_Key_Service), err

}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Product_Package_Service() (SoftLayer_Product_Package_Service, error) {
	service, err := fslc.GetService("SoftLayer_Product_Package")
	return service.(SoftLayer_Product_Package_Service), err
}
func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Product_Order_Service() (SoftLayer_Product_Order_Service, error) {
	service, err := fslc.GetService("SoftLayer_Product_Order")
	return service.(SoftLayer_Product_Order_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Network_Storage_Service() (SoftLayer_Network_Storage_Service, error) {
	service, err := fslc.GetService("SoftLayer_Network_Storage")
	fmt.Println("+++++", service)
	return service.(SoftLayer_Network_Storage_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Billing_Item_Cancellation_Request_Service() (SoftLayer_Billing_Item_Cancellation_Request_Service, error) {
	service, err := fslc.GetService("SoftLayer_Billing_Item_Cancellation_Request")
	return service.(SoftLayer_Billing_Item_Cancellation_Request_Service), err
}

func (fslc *Fake_SoftLayer_Client) GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service() (SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service, error) {
	service, err := fslc.GetService("SoftLayer_Virtual_Guest_Block_Device_Template_Group")
	return service.(SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service), err
}

func (fslc *Fake_SoftLayer_Client) DoRawHttpRequest(path string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	return []byte{}, nil
}

func (fslc *Fake_SoftLayer_Client) DoRawHttpRequestWithObjectMask(path string, masks []string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	return []byte{}, nil
}
func (fslc *Fake_SoftLayer_Client) GenerateRequestBody(templateData interface{}) (*bytes.Buffer, error) {
	return &bytes.Buffer{}, nil
}
func (fslc *Fake_SoftLayer_Client) HasErrors(body map[string]interface{}) error {
	return nil
}

func (fslc *Fake_SoftLayer_Client) CheckForHttpResponseErrors(data []byte) error {
	return nil
}

func (fslc *Fake_SoftLayer_Client) initSoftLayerServices() {
	// fslc.SoftLayerServices["SoftLayer_Account"] = fake_services.NewSoftLayer_Account_Service(fslc)
	fslc.SoftLayerServices["SoftLayer_Virtual_Guest"] = fake_services.NewFake_Virtual_Guest_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Virtual_Disk_Image"] = services.NewSoftLayer_Virtual_Disk_Image_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Security_Ssh_Key"] = services.NewSoftLayer_Security_Ssh_Key_Service(fslc)
	fslc.SoftLayerServices["SoftLayer_Network_Storage"] = fake_services.NewFake_Network_Storage_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Product_Order"] = services.NewSoftLayer_Product_Order_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Product_Package"] = services.NewSoftLayer_Product_Package_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Billing_Item_Cancellation_Request"] = services.NewSoftLayer_Billing_Item_Cancellation_Request_Service(fslc)
	// fslc.SoftLayerServices["SoftLayer_Virtual_Guest_Block_Device_Template_Group"] = services.NewSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service(fslc)
}
