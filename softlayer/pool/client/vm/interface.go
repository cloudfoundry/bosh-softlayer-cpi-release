package vm

import (
	"github.com/go-openapi/runtime"
)

//go:generate counterfeiter -o fakes/fake_softlayer_pool_client.go . SoftLayerPoolClient
type SoftLayerPoolClient interface {
	AddVM(params *AddVMParams) (*AddVMOK, error)
	DeleteVM(params *DeleteVMParams) (*DeleteVMNoContent, error)
	FindVmsByDeployment(params *FindVmsByDeploymentParams) (*FindVmsByDeploymentOK, error)
	FindVmsByFilters(params *FindVmsByFiltersParams) (*FindVmsByFiltersOK, error)
	FindVmsByStates(params *FindVmsByStatesParams) (*FindVmsByStatesOK, error)
	GetVMByCid(params *GetVMByCidParams) (*GetVMByCidOK, error)
	ListVM(params *ListVMParams) (*ListVMOK, error)
	OrderVMByFilter(params *OrderVMByFilterParams) (*OrderVMByFilterOK, error)
	UpdateVM(params *UpdateVMParams) (*UpdateVMOK, error)
	UpdateVMWithState(params *UpdateVMWithStateParams) (*UpdateVMWithStateOK, error)
	SetTransport(transport runtime.ClientTransport)
}
