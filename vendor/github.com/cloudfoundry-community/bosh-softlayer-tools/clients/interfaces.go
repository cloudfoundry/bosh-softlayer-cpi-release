package clients

type BmpClient interface {
	ConfigPath() string

	Info() (InfoResponse, error)
	Bms(deploymentName string) (BmsResponse, error)
	SlPackages() (SlPackagesResponse, error)
	Stemcells() (StemcellsResponse, error)
	SlPackageOptions(packageId string) (SlPackageOptionsResponse, error)
	UpdateState(serverId string, status string) (UpdateStateResponse, error)
	TaskOutput(taskId int, level string) (TaskOutputResponse, error)
	TaskJsonOutput(taskId int, level string) (TaskJsonResponse, error)
	Tasks(latest int) (TasksResponse, error)
	Login(username string, password string) (LoginResponse, error)
	CreateBaremetals(createBaremetalsInfo CreateBaremetalsInfo, dryRun bool) (CreateBaremetalsResponse, error)
	ProvisioningBaremetal(provisioningBaremetalInfo ProvisioningBaremetalInfo) (CreateBaremetalsResponse, error)
}
