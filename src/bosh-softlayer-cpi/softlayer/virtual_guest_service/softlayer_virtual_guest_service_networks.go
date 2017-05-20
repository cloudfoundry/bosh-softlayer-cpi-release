package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"github.com/softlayer/softlayer-go/datatypes"
)

func (vg SoftlayerVirtualGuestService) ConfigureNetworks(id int, networks Networks) (Networks, error) {
	instance, found, err := vg.Find(id)
	if err != nil {
		return networks, err
	}
	if !found {
		return networks, api.NewVMNotFoundError(string(id))
	}

	vg.logger.Info(softlayerVirtualGuestServiceLogTag, "Configuring networks: %#v", networks)
	ubuntu := Softlayer_Ubuntu_Net{
		LinkNamer: NewIndexedNamer(networks),
	}

	componentByNetwork, err := ubuntu.ComponentByNetworkName(instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "ComponentByNetworkName: %#v", componentByNetwork)

	networks, err = ubuntu.NormalizeNetworkDefinitions(networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing network definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Normalized networks: %#v", networks)

	networks, err = ubuntu.NormalizeDynamics(instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing dynamic networks definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Normalized Dynamics: %#v", networks)

	componentByNetwork, err = ubuntu.ComponentByNetworkName(instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "ComponentByNetworkName: %#v", componentByNetwork)

	networks, err = ubuntu.FinalizedNetworkDefinitions(instance, networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Finalizing networks definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finalized network definition: %v", networks)

	return networks, nil
}

func (vg SoftlayerVirtualGuestService) GetVlan(vlanID int, mask string) (datatypes.Network_Vlan, error) {
	return vg.softlayerClient.GetVlan(vlanID, mask)
}
