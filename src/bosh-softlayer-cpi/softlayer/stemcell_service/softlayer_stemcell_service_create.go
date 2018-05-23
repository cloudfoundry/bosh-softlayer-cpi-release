package stemcell

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (s SoftlayerStemcellService) CreateFromTarball(imagePath string, datacenter string, osCode string) (int, error) {
	// Decompress tarball
	defer func() {
		errDefer := s.deleteLocalImageFile(imagePath)
		if errDefer != nil {
			s.logger.Error(softlayerStemcellServiceLogTag, "Delete local VHD file: %s", errDefer)
		}
	}()
	imagePath, err := s.decompressTarBall(imagePath)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Decompress image tarball")
	}

	s.logger.Debug(softlayerStemcellServiceLogTag, "Create a random SoftLayer image name with prefix '%s'", softlayerImageNamePrefix)
	uuidStr, err := s.uuidGen.Generate()
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Generating random prefix")
	}

	// Create a temporary container as image name
	imageName := fmt.Sprintf("%s-%s", softlayerImageNamePrefix, uuidStr)

	defer func() {
		errDefer := s.softlayerClient.DeleteSwiftContainer(imageName)
		if errDefer != nil {
			s.logger.Error(softlayerStemcellServiceLogTag, "Delete SoftLayer Swift container '%s': %s", imageName, errDefer.Error())
		}
	}()

	err = s.softlayerClient.CreateSwiftContainer(imageName)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Create SoftLayer Swift container '%s'", imageName)
	}

	// Upload the image object
	imageFileName := imageName + ".vhd"

	defer func() {
		errDefer := s.softlayerClient.DeleteSwiftLargeObject(imageName, imageFileName)
		if errDefer != nil {
			s.logger.Error(softlayerStemcellServiceLogTag, "Delete SoftLayer Swift large object with '%s/%s': %s", imageName, imageFileName, errDefer.Error())
		}
	}()

	err = s.softlayerClient.UploadSwiftLargeObject(imageName, imageFileName, imagePath)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Create SoftLayer Swift large object with '%s/%s'", imageName, imageFileName)
	}

	// Import
	stemcellId, err := s.softlayerClient.CreateImageFromExternalSource(imageName, "Imported by SL CPI", datacenter, osCode)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Create image from Swift object storage")
	}

	return stemcellId, nil
}

func (s SoftlayerStemcellService) decompressTarBall(source string) (string, error) {
	s.logger.Debug(softlayerStemcellServiceLogTag, "Decompress the file '%s'", source)
	reader, err := os.Open(source)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Open tarball file")
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Create new archive Reader")
	}
	defer archive.Close()

	targetFile := filepath.Join(source + ".vhd")
	s.logger.Debug(softlayerStemcellServiceLogTag, "Create image file '%s'", targetFile)
	writer, err := os.Create(targetFile)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Create image file")
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Copy decompressed image data")
	}

	return targetFile, nil
}

func (s SoftlayerStemcellService) deleteLocalImageFile(filePath string) error {
	s.logger.Debug(softlayerStemcellServiceLogTag, "Delete local image file '%s'", filePath)
	err := os.Remove(filePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Delete image file '%s'", filePath)
	}

	return nil
}
