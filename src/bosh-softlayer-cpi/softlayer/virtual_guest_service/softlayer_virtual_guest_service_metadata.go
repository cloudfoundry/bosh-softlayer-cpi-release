package instance

import (
	"bytes"
	"strconv"
	"time"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	"bosh-softlayer-cpi/api"
)

func (vg SoftlayerVirtualGuestService) SetMetadata(id int, vmMetadata Metadata) error {
	tags, err := vg.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tags from vm metadata")
	}

	var (
		found bool
	)
	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			found, err = vg.softlayerClient.SetTags(id, tags)
			if err != nil {
				return true, bosherr.WrapErrorf(err, "Settings tags on virtualGuest '%d'", id)
			}

			if !found {
				return false, api.NewVMNotFoundError(strconv.Itoa(id))
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, vg.logger.GetBoshLogger())
	vg.logger.ChangeRetryStrategyLogTag(&timeoutRetryStrategy)

	err = timeoutRetryStrategy.Try()
	if err != nil {
		return err
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) extractTagsFromVMMetadata(vmMetadata Metadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	for tagName, tagValue := range vmMetadata {
		if tagName == "name" {
			continue
		}
		tagStringBuffer.WriteString(tagName + ": " + tagValue.(string))
		tagStringBuffer.WriteString(", ")
	}
	tagStringBuffer.Truncate(tagStringBuffer.Len() - 2)

	return tagStringBuffer.String(), nil
}
