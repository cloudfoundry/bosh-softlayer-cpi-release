package stemcell

import (
	"fmt"
	"reflect"
	"unsafe"

	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	"bosh-softlayer-cpi/logger"
	bosl "bosh-softlayer-cpi/softlayer/client"
)

type SoftlayerStemcellService struct {
	softlayerClient bosl.Client
	logger          logger.Logger
}

func NewSoftlayerStemcellService(
	softlayerClient bosl.Client,
	logger logger.Logger,
) SoftlayerStemcellService {
	return SoftlayerStemcellService{
		softlayerClient: softlayerClient,
		logger:          logger,
	}
}

// it's unfriendly to change RetryStrategy(timeoutRetryStrategy).logtag
func (vg SoftlayerStemcellService) changeRetryStrategyLogTag(retryStrategy *boshretry.RetryStrategy) {
	//retryStrategy only refer interface RetryStrategy, so add '*' to get private timeoutRetryStrategy
	pointerVal := reflect.ValueOf(*retryStrategy)
	val := reflect.Indirect(pointerVal)

	logtag := val.FieldByName("logTag")
	ptrToLogTag := unsafe.Pointer(logtag.UnsafeAddr())
	realPtrToLogTag := (*string)(ptrToLogTag)
	serialTagPrefix := fmt.Sprintf("%s:%s", vg.logger.GetSerialTagPrefix(), logtag)
	*realPtrToLogTag = serialTagPrefix
}
