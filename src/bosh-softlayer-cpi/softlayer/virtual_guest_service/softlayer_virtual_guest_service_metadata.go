package instance

import (
	"bytes"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (vg SoftlayerVirtualGuestService) SetMetadata(id int, vmMetadata Metadata) error {
	tags, err := vg.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tags from vm metadata")
	}

	err = vg.softlayerClient.SetTags(id, tags)
	if err != nil {
		return bosherr.WrapErrorf(err, "Settings tags on virtualGuest '%d'", id)
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) extractTagsFromVMMetadata(vmMetadata Metadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	tagStringBuffer.WriteString("deployment" + ":" + vmMetadata["deployment"].(string))
	tagStringBuffer.WriteString(", ")
	tagStringBuffer.WriteString("director" + ":" + vmMetadata["director"].(string))
	tagStringBuffer.WriteString(", ")

	if val, ok := vmMetadata["compiling"]; ok {
		tagStringBuffer.WriteString("compiling" + ":" + val.(string))
	} else {
		tagStringBuffer.WriteString("job" + ":" + vmMetadata["job"].(string))
		tagStringBuffer.WriteString(", ")
		tagStringBuffer.WriteString("index" + ":" + vmMetadata["index"].(string))
	}

	return tagStringBuffer.String(), nil
}
