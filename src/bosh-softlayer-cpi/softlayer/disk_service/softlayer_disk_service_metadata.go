package disk

import (
	"bosh-softlayer-cpi/api"
	"bytes"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"strconv"
)

func (d SoftlayerDiskService) SetMetadata(id int, vmMetadata Metadata) error {
	tags, err := d.generateNotesFromDiskMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "generating notes from disk metadata")
	}

	found, err := d.softlayerClient.SetNotes(id, tags)
	if err != nil {
		return bosherr.WrapErrorf(err, "Settings notes on network storage '%d'", id)
	}

	if !found {
		return api.NewDiskNotFoundError(strconv.Itoa(id), false)
	}

	return nil
}

func (d SoftlayerDiskService) generateNotesFromDiskMetadata(vmMetadata Metadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	if val, ok := vmMetadata["director"]; ok {
		tagStringBuffer.WriteString("director" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["deployment"]; ok {
		tagStringBuffer.WriteString("deployment" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["instance_id"]; ok {
		tagStringBuffer.WriteString("instance_id" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["job"]; ok {
		tagStringBuffer.WriteString("job" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["instance_index"]; ok {
		tagStringBuffer.WriteString("instance_index" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["instance_name"]; ok {
		tagStringBuffer.WriteString("instance_name" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}

	if val, ok := vmMetadata["attached_at"]; ok {
		tagStringBuffer.WriteString("attached_at" + ": " + val.(string))
		tagStringBuffer.WriteString(", ")
	}
	tagStringBuffer.Truncate(tagStringBuffer.Len() - 2)

	return tagStringBuffer.String(), nil
}
