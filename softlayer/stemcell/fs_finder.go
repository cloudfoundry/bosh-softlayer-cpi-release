package stemcell

import (
  // "fmt"
  // "strconv"
  
  sl "github.com/maximilien/softlayer-go/softlayer"
  boshlog "github.com/cloudfoundry/bosh-agent/logger"
  //   bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type FSFinder struct {
	client sl.Client
	logger boshlog.Logger
}

func NewFSFinder(client sl.Client, logger boshlog.Logger) FSFinder {
	return FSFinder{client: client, logger: logger}
}

func (f FSFinder) Find(id string) (Stemcell, bool, error) {
  // stemcellImageId, err := strconv.Atoi(id)
  // if err != nil {
  //   return nil, false, bosherr.WrapError(err, "Converting stemcell id to int")
  // }
  //
  // accountService, err := f.client.GetSoftLayer_Account_Service()
  // if err != nil {
  //   return nil, false, bosherr.WrapError(err, "Getting virtual guest service")
  // }
  //
  // virtualDiskImages, err := accountService.GetVirtualDiskImages()
  // if err != nil {
  //   return nil, false, bosherr.WrapError(err, "get virtual disk images")
  // }
  
  return NewFSStemcell("200150", f.logger), true, nil

  //   for _, vdImage := range virtualDiskImages {
  //     // fmt.Printf("===>vdImage.Id: %d\n", vdImage.Id) //DEBUG
  //     // fmt.Printf("===>vdImage.Uuid: %d\n", vdImage.Uuid) //DEBUG
  //     if vdImage.Id == stemcellImageId {
  //       return NewFSStemcell(id, f.logger), true, nil
  //     }
  //   }
  //
  // return nil, false, nil
}
