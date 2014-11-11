package disk

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	// boshsys "github.com/cloudfoundry/bosh-agent/system"

	"errors"
	"fmt"
	sl "github.com/maximilien/softlayer-go/softlayer"
	"reflect"
)

const (
	iscsiCreatorLogTag = "SoftLayerIscsiCreator"
	DEFAULT_LOCATION   = "138124" // ToDo: Set default data-center: dal05, in the future, we need to use service to randomly pick up a data center
)

type SoftLayerCreator struct {
	softlayerClient sl.Client
	logger          boshlog.Logger
}

func NewSoftLayerCreator(
	client sl.Client,
	logger boshlog.Logger,
) SoftLayerCreator {
	return SoftLayerCreator{
		softlayerClient: client,
		logger:          logger,
	}
}

func (slc SoftLayerCreator) Create(size int, virtualGuestId int) (Disk, error) {
	slc.logger.Debug(iscsiCreatorLogTag, "Creating disk of size '%d'", size)

	if size <= 0 {
		return IscsiDisk{}, errors.New(fmt.Sprintf("Illegal disk size: %d", size))
	}

	networkStorageService, err := slc.softlayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return IscsiDisk{}, bosherr.WrapError(err, "Can not get SoftLayer_Network_Storage_Service from softlayer-go client")
	}

	virtualGuestService, err := slc.softlayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return IscsiDisk{}, bosherr.WrapError(err, "Can not get SoftLayer_Virtual_Guest_Service from softlayer-go client")
	}

	fmt.Println("--------", reflect.TypeOf(slc.softlayerClient))

	virtualGuest, err := virtualGuestService.GetObject(virtualGuestId)
	if err != nil {
		return IscsiDisk{}, bosherr.WrapError(err, "Can not get virtual guest for id %s", virtualGuestId)
	}
	fmt.Println("virtualGuest: ", virtualGuest)
	// location := virtualGuest.Location.Id
	location := 123
	// fmt.Println("The location is:", location)
	volume, err := networkStorageService.CreateIscsiVolume(size, string(location))
	if err != nil {
		return IscsiDisk{}, bosherr.WrapError(err, "Can not create iSCSI volume")
	}

	return IscsiDisk{id: volume.Id, logger: slc.logger}, nil
}
