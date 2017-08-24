package instance

import (
	boslc "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"github.com/softlayer/softlayer-go/datatypes"
	"strconv"
)

func (vg SoftlayerVirtualGuestService) ConfigureNetworks(id int, networks Networks) (Networks, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest '%d' ", id)
	instance, found, err := vg.softlayerClient.GetInstance(id, boslc.INSTANCE_DETAIL_MASK)
	if err != nil {
		return networks, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with id '%d'", id)
	}

	if !found {
		return networks, api.NewVMNotFoundError(strconv.Itoa(id))
	}

	vg.logger.Info(softlayerVirtualGuestServiceLogTag, "Configuring networks: %+v", networks)
	ubuntu := Softlayer_Ubuntu_Net{
		LinkNamer: NewIndexedNamer(networks),
	}

	componentByNetwork, err := ubuntu.ComponentByNetworkName(*instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "ComponentByNetworkName: %+v", componentByNetwork)

	networks, err = ubuntu.NormalizeNetworkDefinitions(networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing network definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Normalized networks: %+v", networks)

	networks, err = ubuntu.NormalizeDynamics(*instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing dynamic networks definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Normalized Dynamics: %+v", networks)

	componentByNetwork, err = ubuntu.ComponentByNetworkName(*instance, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "ComponentByNetworkName: %+v", componentByNetwork)

	networks, err = ubuntu.FinalizedNetworkDefinitions(*instance, networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Finalizing networks definitions")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finalized network definition: %+v", networks)

	return networks, nil
}

func (vg SoftlayerVirtualGuestService) GetVlan(vlanID int, mask string) (*datatypes.Network_Vlan, error) {
	vlan, found, err := vg.softlayerClient.GetVlan(vlanID, mask)
	if err != nil {
		return &datatypes.Network_Vlan{}, bosherr.WrapErrorf(err, "Getting vlan details with id '%d'", vlanID)
	}

	if !found {
		return &datatypes.Network_Vlan{}, bosherr.Errorf("Failed to get vlan details with id '%d'", vlanID)
	}

	return vlan, nil
}

func (vg SoftlayerVirtualGuestService) GetSubnet(subnetID int, mask string) (*datatypes.Network_Subnet, error) {
	subnet, found, err := vg.softlayerClient.GetSubnet(subnetID, mask)
	if err != nil {
		return &datatypes.Network_Subnet{}, bosherr.WrapErrorf(err, "Getting subnet details with id '%d'", subnetID)
	}

	if !found {
		return &datatypes.Network_Subnet{}, bosherr.Errorf("Failed to get subnet details with id '%d'", subnetID)
	}

	return subnet, nil
}
