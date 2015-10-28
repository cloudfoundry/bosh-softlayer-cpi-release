package vm

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	datatypes "github.com/maximilien/softlayer-go/data_types"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
)

type SoftlayerFileService interface {
	Upload(string, []byte) error
	Download(string) ([]byte, error)
}

type softlayerFileService struct {
	sshClient    util.SshClient
	virtualGuest datatypes.SoftLayer_Virtual_Guest
	logger       boshlog.Logger
	logTag       string
}

func NewSoftlayerFileService(sshClient util.SshClient, virtualGuest datatypes.SoftLayer_Virtual_Guest, logger boshlog.Logger) SoftlayerFileService {
	return &softlayerFileService{
		sshClient:    sshClient,
		virtualGuest: virtualGuest,
		logger:       logger,
		logTag:       "softlayerFileService",
	}
}

func (s *softlayerFileService) Download(sourcePath string) ([]byte, error) {
	sourceFileName := filepath.Base(sourcePath)
	tmpFilePath := filepath.Join("/tmp", sourceFileName)

	s.logger.Debug(s.logTag, "Downloading file at %s", sourcePath)

	s.sshClient.DownloadFile(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, sourcePath, tmpFilePath)

	fileSystem := boshsys.NewOsFileSystemWithStrictTempRoot(s.logger)
	contents, err := fileSystem.ReadFile(tmpFilePath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading from %s", tmpFilePath)
	}

	s.logger.Debug(s.logTag, "Read user data '%#v'", contents)

	return []byte(contents), nil
}

func (s *softlayerFileService) Upload(destinationPath string, contents []byte) error {
	s.logger.Debug(s.logTag, "Uploading file to %s", destinationPath)

	destinationFileName := filepath.Base(destinationPath)
	tmpFilePath := filepath.Join("/tmp", destinationFileName)

	fileSystem := boshsys.NewOsFileSystemWithStrictTempRoot(s.logger)

	err := fileSystem.WriteFile(tmpFilePath, contents)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing to %s", tmpFilePath)
	}

	err = s.sshClient.UploadFile(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, tmpFilePath, destinationPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Moving temporary file to destination '%s'", destinationPath)
	}

	return nil
}

func (s *softlayerFileService) getRootPassword(virtualGuest datatypes.SoftLayer_Virtual_Guest) string {
	passwords := virtualGuest.OperatingSystem.Passwords
	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return ""
}
