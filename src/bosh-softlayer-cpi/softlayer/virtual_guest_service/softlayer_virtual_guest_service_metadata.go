package instance

import (
	"bytes"
	"regexp"
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
	err = vg.logger.ChangeRetryStrategyLogTag(&timeoutRetryStrategy)
	if err != nil {
		return err
	}

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

		// https://softlayer.github.io/reference/services/SoftLayer_Tag/setTags
		// The characters permitted are A-Z, 0-9, whitespace, _ (underscore), - (hypen), . (period), and : (colon).
		reg, err := regexp.Compile(`[^\w \-.:]+`)
		if err != nil {
			return "", bosherr.WrapError(err, "There is a problem with your regexp: '[^\\w \\-.:]+'. That is used to strips out all invalid characters")
		}
		cleanTagString := reg.ReplaceAllString(tagValue.(string), "")

		regColon, err := regexp.Compile(`[:]+`)
		if err != nil {
			return "", bosherr.WrapError(err, "There is a problem with your regexp: '[:]+'. That is used to strips out all ':' characters")
		}
		convertedTagString := regColon.ReplaceAllString(cleanTagString, "-")

		_, err = tagStringBuffer.WriteString(tagName + ":" + convertedTagString)
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}
	tagStringBuffer.Truncate(tagStringBuffer.Len() - 1)

	return tagStringBuffer.String(), nil
}
