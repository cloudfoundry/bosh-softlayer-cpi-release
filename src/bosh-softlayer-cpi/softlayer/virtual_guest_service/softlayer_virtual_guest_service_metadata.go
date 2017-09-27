package instance

import (
	"bosh-softlayer-cpi/api"
	"bytes"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"strconv"
)

func (vg SoftlayerVirtualGuestService) SetMetadata(id int, vmMetadata Metadata) error {
	tags, err := vg.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tags from vm metadata")
	}

	found, err := vg.softlayerClient.SetTags(id, tags)
	if err != nil {
		return bosherr.WrapErrorf(err, "Settings tags on virtualGuest '%d'", id)
	}

	if !found {
		return api.NewVMNotFoundError(strconv.Itoa(id))
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) extractTagsFromVMMetadata(vmMetadata Metadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	if val, ok := vmMetadata["deployment"]; ok {
		tagStringBuffer.WriteString("deployment" + ":" + val.(string))
	}
	if val, ok := vmMetadata["director"]; ok {
		tagStringBuffer.WriteString(", ")
		tagStringBuffer.WriteString("director" + ":" + val.(string))
	}

	if val, ok := vmMetadata["compiling"]; ok {
		tagStringBuffer.WriteString(", ")
		tagStringBuffer.WriteString("compiling" + ":" + val.(string))
	} else {
		if val, ok := vmMetadata["job"]; ok {
			tagStringBuffer.WriteString(", ")
			tagStringBuffer.WriteString("job" + ":" + val.(string))
		}
		if val, ok := vmMetadata["index"]; ok {
			tagStringBuffer.WriteString(", ")
			tagStringBuffer.WriteString("index" + ":" + val.(string))
		}
	}

	return tagStringBuffer.String(), nil
}
