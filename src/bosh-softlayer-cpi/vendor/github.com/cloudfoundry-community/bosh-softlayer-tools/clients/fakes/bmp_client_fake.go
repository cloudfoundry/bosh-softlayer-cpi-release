package fakes

import (
	clients "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
)

type FakeBmpClient struct {
	Username string
	Password string
	Url      string

	ConfigPathString string

	InfoResponsesIndex int
	InfoResponses      []clients.InfoResponse
	InfoResponse       clients.InfoResponse
	InfoErr            error

	BmsResponsesIndex int
	BmsResponses      []clients.BmsResponse
	BmsResponse       clients.BmsResponse
	BmsErr            error

	SlPackagesResponsesIndex int
	SlPackagesResponses      []clients.SlPackagesResponse
	SlPackagesResponse       clients.SlPackagesResponse
	SlPackagesErr            error

	StemcellsResponsesIndex int
	StemcellsResponses      []clients.StemcellsResponse
	StemcellsResponse       clients.StemcellsResponse
	StemcellsErr            error

	SlPackageOptionsResponsesIndex int
	SlPackageOptionsResponses      []clients.SlPackageOptionsResponse
	SlPackageOptionsResponse       clients.SlPackageOptionsResponse
	SlPackageOptionsErr            error

	TasksResponsesIndex int
	TasksResponses      []clients.TasksResponse
	TasksResponse       clients.TasksResponse
	TasksErr            error

	TaskOutputResponsesIndex int
	TaskOutputResponses      []clients.TaskOutputResponse
	TaskOutputResponse       clients.TaskOutputResponse
	TaskOutputErr            error

	TaskJsonResponsesIndex int
	TaskJsonResponses      []clients.TaskJsonResponse
	TaskJsonResponse       clients.TaskJsonResponse
	TaskJsonOutputErr      error

	UpdateStateResponsesIndex int
	UpdateStateResponses      []clients.UpdateStateResponse
	UpdateStateResponse       clients.UpdateStateResponse
	UpdateStateErr            error

	LoginResponse clients.LoginResponse
	LoginErr      error

	CreateBaremetalsResponsesIndex int
	CreateBaremetalsResponses      []clients.CreateBaremetalsResponse
	CreateBaremetalsResponse       clients.CreateBaremetalsResponse
	CreateBaremetalsErr            error

	ProvisioningBaremetalResponsesIndex int
	ProvisioningBaremetalResponses      []clients.CreateBaremetalsResponse
	ProvisioningBaremetalResponse       clients.CreateBaremetalsResponse
	ProvisioningBaremetalErr            error
}

func NewFakeBmpClient(username, password, url string, configPath string) *FakeBmpClient {
	return &FakeBmpClient{
		Username:         username,
		Password:         password,
		Url:              url,
		ConfigPathString: configPath,
	}
}

func (bc *FakeBmpClient) ConfigPath() string {
	return bc.ConfigPathString
}

func (bc *FakeBmpClient) Info() (clients.InfoResponse, error) {
	if bc.InfoErr != nil {
		return bc.InfoResponse, bc.InfoErr
	}

	if len(bc.InfoResponses) == 0 {
		return bc.InfoResponse, bc.InfoErr
	} else {
		bc.InfoResponsesIndex = bc.InfoResponsesIndex + 1
		return bc.InfoResponses[bc.InfoResponsesIndex-1], bc.InfoErr
	}
}

func (bc *FakeBmpClient) Bms(deploymentName string) (clients.BmsResponse, error) {
	if bc.BmsErr != nil {
		return bc.BmsResponse, bc.BmsErr
	}

	if len(bc.BmsResponses) == 0 {
		return bc.BmsResponse, bc.BmsErr
	} else {
		bc.BmsResponsesIndex = bc.BmsResponsesIndex + 1
		return bc.BmsResponses[bc.BmsResponsesIndex-1], bc.BmsErr
	}
}

func (bc *FakeBmpClient) SlPackages() (clients.SlPackagesResponse, error) {
	if bc.SlPackagesErr != nil {
		return bc.SlPackagesResponse, bc.SlPackagesErr
	}

	if len(bc.SlPackagesResponses) == 0 {
		return bc.SlPackagesResponse, bc.SlPackagesErr
	} else {
		bc.SlPackagesResponsesIndex = bc.SlPackagesResponsesIndex + 1
		return bc.SlPackagesResponses[bc.SlPackagesResponsesIndex-1], bc.SlPackagesErr
	}
}

func (bc *FakeBmpClient) Stemcells() (clients.StemcellsResponse, error) {
	if bc.StemcellsErr != nil {
		return bc.StemcellsResponse, bc.StemcellsErr
	}

	if len(bc.StemcellsResponses) == 0 {
		return bc.StemcellsResponse, bc.StemcellsErr
	} else {
		bc.StemcellsResponsesIndex = bc.StemcellsResponsesIndex + 1
		return bc.StemcellsResponses[bc.StemcellsResponsesIndex-1], bc.StemcellsErr
	}
}

func (bc *FakeBmpClient) SlPackageOptions(packageId string) (clients.SlPackageOptionsResponse, error) {
	if bc.SlPackageOptionsErr != nil {
		return bc.SlPackageOptionsResponse, bc.SlPackageOptionsErr
	}

	if len(bc.SlPackageOptionsResponses) == 0 {
		return bc.SlPackageOptionsResponse, bc.SlPackageOptionsErr
	} else {
		bc.SlPackageOptionsResponsesIndex = bc.SlPackageOptionsResponsesIndex + 1
		return bc.SlPackageOptionsResponses[bc.SlPackageOptionsResponsesIndex-1], bc.SlPackageOptionsErr
	}
}

func (bc *FakeBmpClient) Tasks(latest int) (clients.TasksResponse, error) {
	if bc.TasksErr != nil {
		return bc.TasksResponse, bc.TasksErr
	}

	if len(bc.TasksResponses) == 0 {
		return bc.TasksResponse, bc.TasksErr
	} else {
		bc.TasksResponsesIndex = bc.TasksResponsesIndex + 1
		return bc.TasksResponses[bc.TasksResponsesIndex-1], bc.TasksErr
	}
}

func (bc *FakeBmpClient) TaskOutput(taskID int, level string) (clients.TaskOutputResponse, error) {
	if bc.TaskOutputErr != nil {
		return bc.TaskOutputResponse, bc.TaskOutputErr
	}

	if len(bc.TaskOutputResponses) == 0 {
		return bc.TaskOutputResponse, bc.TaskOutputErr
	} else {
		bc.TaskOutputResponsesIndex = bc.TaskOutputResponsesIndex + 1
		return bc.TaskOutputResponses[bc.TaskOutputResponsesIndex-1], bc.TaskOutputErr
	}
}

func (bc *FakeBmpClient) TaskJsonOutput(taskID int, level string) (clients.TaskJsonResponse, error) {
	if bc.TaskJsonOutputErr != nil {
		return bc.TaskJsonResponse, bc.TaskJsonOutputErr
	}

	if len(bc.TaskJsonResponses) == 0 {
		return bc.TaskJsonResponse, bc.TaskJsonOutputErr
	} else {
		bc.TaskJsonResponsesIndex = bc.TaskJsonResponsesIndex + 1
		return bc.TaskJsonResponses[bc.TaskJsonResponsesIndex-1], bc.TaskJsonOutputErr
	}
}

func (bc *FakeBmpClient) UpdateState(serverId string, status string) (clients.UpdateStateResponse, error) {
	if bc.UpdateStateErr != nil {
		return bc.UpdateStateResponse, bc.UpdateStateErr
	}

	if len(bc.TaskJsonResponses) == 0 {
		return bc.UpdateStateResponse, bc.UpdateStateErr
	} else {
		bc.UpdateStateResponsesIndex = bc.UpdateStateResponsesIndex + 1
		return bc.UpdateStateResponses[bc.UpdateStateResponsesIndex-1], bc.UpdateStateErr
	}
}

func (bc *FakeBmpClient) Login(username string, password string) (clients.LoginResponse, error) {
	return bc.LoginResponse, bc.LoginErr
}

func (bc *FakeBmpClient) CreateBaremetals(CreateBaremetalsInfo clients.CreateBaremetalsInfo, DryRun bool) (clients.CreateBaremetalsResponse, error) {
	if bc.CreateBaremetalsErr != nil {
		return bc.CreateBaremetalsResponse, bc.CreateBaremetalsErr
	}

	if len(bc.CreateBaremetalsResponses) == 0 {
		return bc.CreateBaremetalsResponse, bc.CreateBaremetalsErr
	} else {
		bc.CreateBaremetalsResponsesIndex = bc.CreateBaremetalsResponsesIndex + 1
		return bc.CreateBaremetalsResponses[bc.CreateBaremetalsResponsesIndex-1], bc.CreateBaremetalsErr
	}
}

func (bc *FakeBmpClient) ProvisioningBaremetal(provisioningBaremetalInfo clients.ProvisioningBaremetalInfo) (clients.CreateBaremetalsResponse, error) {
	if bc.ProvisioningBaremetalErr != nil {
		return bc.ProvisioningBaremetalResponse, bc.ProvisioningBaremetalErr
	}

	if len(bc.ProvisioningBaremetalResponses) == 0 {
		return bc.ProvisioningBaremetalResponse, bc.ProvisioningBaremetalErr
	} else {
		bc.ProvisioningBaremetalResponsesIndex = bc.ProvisioningBaremetalResponsesIndex + 1
		return bc.ProvisioningBaremetalResponses[bc.ProvisioningBaremetalResponsesIndex-1], bc.ProvisioningBaremetalErr
	}
}
