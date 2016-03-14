package vm

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	datatypes "github.com/maximilien/softlayer-go/data_types"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type SoftlayerFileService interface {
	Upload(string, []byte) error
	Download(string) ([]byte, error)
}

type softlayerFileService struct {
	sshClient     util.SshClient
	virtualGuest  datatypes.SoftLayer_Virtual_Guest
	logger        boshlog.Logger
	logTag        string
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
}

func NewSoftlayerFileService(sshClient util.SshClient, virtualGuest datatypes.SoftLayer_Virtual_Guest, logger boshlog.Logger, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem) SoftlayerFileService {
	return &softlayerFileService{
		sshClient:     sshClient,
		virtualGuest:  virtualGuest,
		logger:        logger,
		logTag:        "softlayerFileService",
		uuidGenerator: uuidGenerator,
		fs:            fs,
	}
}

func (s *softlayerFileService) Download(sourcePath string) ([]byte, error) {
	s.logger.Debug(s.logTag, "Downloading file at %s", sourcePath)

	tmpDirUUID, err := s.uuidGenerator.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating uuid for temp file")
	}
	tmpDir, err := s.fs.TempDir(tmpDirUUID)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting temp dir for downloading user_data.json")
	}

	defer s.fs.RemoveAll(tmpDir)

	sourceFileName := filepath.Base(sourcePath)
	tmpFilePath := filepath.Join(tmpDir, sourceFileName)

	s.sshClient.DownloadFile(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, sourcePath, tmpFilePath)

	contents, err := s.fs.ReadFile(tmpFilePath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading from %s", tmpFilePath)
	}

	s.logger.Debug(s.logTag, "Read user data '%#v'", contents)

	return []byte(contents), nil
}

func (s *softlayerFileService) Upload(destinationPath string, contents []byte) error {
	s.logger.Debug(s.logTag, "Uploading file to %s", destinationPath)

	tmpDirUUID, err := s.uuidGenerator.Generate()
	if err != nil {
		return bosherr.WrapError(err, "Generating uuid for temp file")
	}
	tmpDir, err := s.fs.TempDir(tmpDirUUID)
	if err != nil {
		return bosherr.WrapError(err, "Getting temp dir for uploading user_data.json")
	}

	defer s.fs.RemoveAll(tmpDir)

	sourceFileName := filepath.Base(destinationPath)
	tmpFilePath := filepath.Join(tmpDir, sourceFileName)

	err = s.fs.WriteFile(tmpFilePath, contents)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing to %s", tmpFilePath)
	}

	err = s.sshClient.UploadFile(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, tmpFilePath, destinationPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Uploading temporary file to destination '%s'", destinationPath)
	}

	return nil
}

// private method
func (s *softlayerFileService) getRootPassword(virtualGuest datatypes.SoftLayer_Virtual_Guest) string {
	passwords := virtualGuest.OperatingSystem.Passwords
	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return ""
}
