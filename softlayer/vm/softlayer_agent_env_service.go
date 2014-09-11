package vm

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
)

const (
	softLayerAgentEnvServiceLogTag = "softLayerAgentEnvService"

	softLayerAgentEnvServiceSettingsFileName  = "softlayer-cpi-agent-env.json"
	softLayerAgentEnvServiceTmpSettingsPath   = "/tmp/" + softLayerAgentEnvServiceSettingsFileName
	softLayerAgentEnvServiceFinalSettingsPath = "/var/vcap/bosh/" + softLayerAgentEnvServiceSettingsFileName
)

type SoftLayerAgentEnvService struct {
	container wrdn.Container
	logger    boshlog.Logger
}

func NewSoftLayerAgentEnvService(
	container wrdn.Container,
	logger boshlog.Logger,
) SoftLayerAgentEnvService {
	return SoftLayerAgentEnvService{
		container: container,
		logger:    logger,
	}
}

func (s SoftLayerAgentEnvService) Fetch() (AgentEnv, error) {
	// Copy settings file to a temporary directory
	// so that tar (running as vcap) has permission to readdir.
	// (/var/vcap/bosh is owned by root.)
	script := fmt.Sprintf(
		"cp %s %s && chown vcap:vcap %s",
		softLayerAgentEnvServiceFinalSettingsPath,
		softLayerAgentEnvServiceTmpSettingsPath,
		softLayerAgentEnvServiceTmpSettingsPath,
	)

	err := s.runPrivilegedScript(script)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Running copy json settings file script")
	}

	streamOut, err := s.container.StreamOut(softLayerAgentEnvServiceTmpSettingsPath)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Streaming out json settings")
	}

	return s.unmarshalAgentEnv(streamOut)
}

func (s SoftLayerAgentEnvService) Update(agentEnv AgentEnv) error {
	agentEnvStream, err := s.marshalAgentEnv(agentEnv, softLayerAgentEnvServiceSettingsFileName)
	if err != nil {
		return bosherr.WrapError(err, "Making json settings stream")
	}

	// Stream in settings file to a temporary directory
	// so that tar (running as vcap) has permission to unpack into dir.
	// Do not directly write to /var/vcap/bosh/settings.json.
	// That file path is an implementation detail of BOSH Agent.
	err = s.container.StreamIn("/tmp/", agentEnvStream)
	if err != nil {
		return bosherr.WrapError(err, "Streaming in json settings")
	}

	// Move settings file to its final location
	script := fmt.Sprintf(
		"mv %s %s",
		softLayerAgentEnvServiceTmpSettingsPath,
		softLayerAgentEnvServiceFinalSettingsPath,
	)

	err = s.runPrivilegedScript(script)
	if err != nil {
		return bosherr.WrapError(err, "Running move json settings file script")
	}

	return nil
}

func (s SoftLayerAgentEnvService) unmarshalAgentEnv(agentEnvStream io.Reader) (AgentEnv, error) {
	var agentEnv AgentEnv

	tarReader := tar.NewReader(agentEnvStream)

	_, err := tarReader.Next()
	if err != nil {
		return agentEnv, bosherr.WrapError(err, "Reading tar header for agent env")
	}

	err = json.NewDecoder(tarReader).Decode(&agentEnv)
	if err != nil {
		return agentEnv, bosherr.WrapError(err, "Reading agent env from tar")
	}

	s.logger.Debug(softLayerAgentEnvServiceLogTag, "Unmarshalled agent env: %#v", agentEnv)

	return agentEnv, nil
}

func (s SoftLayerAgentEnvService) marshalAgentEnv(agentEnv AgentEnv, fileName string) (io.Reader, error) {
	s.logger.Debug(softLayerAgentEnvServiceLogTag, "Marshalling agent env: %#v", agentEnv)

	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		return nil, bosherr.WrapError(err, "Marshalling agent env")
	}

	tarBytes := &bytes.Buffer{}

	tarWriter := tar.NewWriter(tarBytes)

	fileHeader := &tar.Header{
		Name: fileName,
		Size: int64(len(jsonBytes)),
		Mode: 0640,
	}

	err = tarWriter.WriteHeader(fileHeader)
	if err != nil {
		return nil, bosherr.WrapError(err, "Writing tar header for agent env")
	}

	_, err = tarWriter.Write(jsonBytes)
	if err != nil {
		return nil, bosherr.WrapError(err, "Writing agent env to tar")
	}

	err = tarWriter.Close()
	if err != nil {
		return nil, bosherr.WrapError(err, "Closing tar writer")
	}

	return tarBytes, nil
}

func (s SoftLayerAgentEnvService) runPrivilegedScript(script string) error {
	processSpec := wrdn.ProcessSpec{
		Path: "bash",
		Args: []string{"-c", script},

		Privileged: true,
	}

	process, err := s.container.Run(processSpec, wrdn.ProcessIO{})
	if err != nil {
		return bosherr.WrapError(err, "Running script")
	}

	exitCode, err := process.Wait()
	if err != nil {
		return bosherr.WrapError(err, "Waiting for script")
	}

	if exitCode != 0 {
		return bosherr.New("Script exited with non-0 exit code")
	}

	return nil
}
