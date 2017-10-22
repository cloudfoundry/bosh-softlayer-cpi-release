package instance

import (
	"fmt"
	"reflect"
	"unsafe"

	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

const rootUser = "root"
const softlayerVirtualGuestServiceLogTag = "SoftlayerVirtualGuestService"

type SoftlayerVirtualGuestService struct {
	softlayerClient bosl.Client
	uuidGen         boshuuid.Generator
	logger          logger.Logger
}

func NewSoftLayerVirtualGuestService(
	softlayerClient bosl.Client,
	uuidGen boshuuid.Generator,
	logger logger.Logger,
) SoftlayerVirtualGuestService {
	return SoftlayerVirtualGuestService{
		softlayerClient: softlayerClient,
		uuidGen:         uuidGen,
		logger:          logger,
	}
}

type Mount struct {
	PartitionPath string
	MountPoint    string
}

// it's unfriendly to change RetryStrategy().logtag
func (vg SoftlayerVirtualGuestService) changeRetryStrategyLogTag(retryStrategy *boshretry.RetryStrategy) {
	//retryStrategy only refer interface RetryStrategy, so add '*' to get private timeoutRetryStrategy
	pointerVal := reflect.ValueOf(*retryStrategy)
	val := reflect.Indirect(pointerVal)

	logtag := val.FieldByName("logTag")
	ptrToLogTag := unsafe.Pointer(logtag.UnsafeAddr())
	realPtrToLogTag := (*string)(ptrToLogTag)
	serialTagPrefix := fmt.Sprintf("%s:%s", vg.logger.GetSerialTagPrefix(), logtag)
	*realPtrToLogTag = serialTagPrefix
}
