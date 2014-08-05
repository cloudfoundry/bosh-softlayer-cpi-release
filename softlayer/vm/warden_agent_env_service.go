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
	wardenAgentEnvServiceLogTag = "WardenAgentEnvService"

	wardenAgentEnvServiceSettingsFileName  = "warden-cpi-agent-env.json"
	wardenAgentEnvServiceTmpSettingsPath   = "/tmp/" + wardenAgentEnvServiceSettingsFileName
	wardenAgentEnvServiceFinalSettingsPath = "/var/vcap/bosh/" + wardenAgentEnvServiceSettingsFileName
)

type WardenAgentEnvService struct {
	container wrdn.Container
	logger    boshlog.Logger
}

func NewWardenAgentEnvService(
	container wrdn.Container,
	logger boshlog.Logger,
) WardenAgentEnvService {
	return WardenAgentEnvService{
		container: container,
		logger:    logger,
	}
}

func (s WardenAgentEnvService) Fetch() (AgentEnv, error) {
	// Copy settings file to a temporary directory
	// so that tar (running as vcap) has permission to readdir.
	// (/var/vcap/bosh is owned by root.)
	script := fmt.Sprintf(
		"cp %s %s && chown vcap:vcap %s",
		wardenAgentEnvServiceFinalSettingsPath,
		wardenAgentEnvServiceTmpSettingsPath,
		wardenAgentEnvServiceTmpSettingsPath,
	)

	err := s.runPrivilegedScript(script)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Running copy json settings file script")
	}

	streamOut, err := s.container.StreamOut(wardenAgentEnvServiceTmpSettingsPath)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Streaming out json settings")
	}

	return s.unmarshalAgentEnv(streamOut)
}

func (s WardenAgentEnvService) Update(agentEnv AgentEnv) error {
	agentEnvStream, err := s.marshalAgentEnv(agentEnv, wardenAgentEnvServiceSettingsFileName)
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
		wardenAgentEnvServiceTmpSettingsPath,
		wardenAgentEnvServiceFinalSettingsPath,
	)

	err = s.runPrivilegedScript(script)
	if err != nil {
		return bosherr.WrapError(err, "Running move json settings file script")
	}

	return nil
}

func (s WardenAgentEnvService) unmarshalAgentEnv(agentEnvStream io.Reader) (AgentEnv, error) {
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

	s.logger.Debug(wardenAgentEnvServiceLogTag, "Unmarshalled agent env: %#v", agentEnv)

	return agentEnv, nil
}

func (s WardenAgentEnvService) marshalAgentEnv(agentEnv AgentEnv, fileName string) (io.Reader, error) {
	s.logger.Debug(wardenAgentEnvServiceLogTag, "Marshalling agent env: %#v", agentEnv)

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

func (s WardenAgentEnvService) runPrivilegedScript(script string) error {
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
