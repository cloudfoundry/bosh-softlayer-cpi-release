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
	var (
		tagStringBuffer bytes.Buffer
		err             error
	)
	if val, ok := vmMetadata["director"]; ok {
		_, err = tagStringBuffer.WriteString("director" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["deployment"]; ok {
		_, err = tagStringBuffer.WriteString("deployment" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["instance_id"]; ok {
		_, err = tagStringBuffer.WriteString("instance_id" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["job"]; ok {
		_, err = tagStringBuffer.WriteString("job" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["instance_index"]; ok {
		_, err = tagStringBuffer.WriteString("instance_index" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["instance_name"]; ok {
		_, err = tagStringBuffer.WriteString("instance_name" + ":" + val.(string))
		if err != nil {
			return tagStringBuffer.String(), err
		}
		_, err = tagStringBuffer.WriteString(",")
		if err != nil {
			return tagStringBuffer.String(), err
		}
	}

	if val, ok := vmMetadata["attached_at"]; ok {
		_, err = tagStringBuffer.WriteString("attached_at" + ":" + val.(string))
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
