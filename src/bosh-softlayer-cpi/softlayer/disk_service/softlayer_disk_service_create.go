package disk

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"math"
)

func (d SoftlayerDiskService) Create(size int, iops int, location string) (int, error) {
	d.logger.Debug(softlayerDiskServiceLogTag, "Creating disk of size '%d'", size)
	volume, err := d.softlayerClient.CreateVolume(location, d.getSoftLayerDiskSize(size), iops)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Creating volume with size '%d', iops '%d', location `%sd`", d.getSoftLayerDiskSize(size), iops, location)
	}

	return *volume.Id, nil
}

func (d SoftlayerDiskService) getSoftLayerDiskSize(size int) int {
	// Sizes and IOPS ranges: http://knowledgelayer.softlayer.com/learning/performance-storage-concepts
	sizeArray := []int{20, 40, 80, 100, 250, 500, 1000, 2000, 4000, 8000, 12000}

	for _, value := range sizeArray {
		if sizeGb := float64(size) / float64(1024); int(math.Ceil(sizeGb)) <= value {
			return value
		}
	}
	return 12000
}
