package test_helpers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"text/template"
)

type CpiTemplate struct {
	ID             string
	DirectorUuid   string
	Tag_deployment string
	Tag_compiling  string
	Datacenter     string
	VMID           string
	DiskID         string
}

type ConfigTemplate struct {
	Username string
	ApiKey   string
}

const templatePath = "../test_fixtures/cpi_methods"

func RunCpi(rootCpiPath string, configPath string, jsonPayload string) ([]byte, error) {
	cpiPath := filepath.Join(rootCpiPath, "out/cpi")

	cmd := exec.Command(cpiPath, "-configPath", configPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return []byte{}, err
	}

	_, err = stdin.Write([]byte(jsonPayload))
	if err != nil {
		return []byte{}, err
	}

	err = stdin.Close()
	if err != nil {
		return []byte{}, err
	}

	output, err := cmd.Output()
	if err != nil {
		return []byte{}, err
	}

	return output, nil
}

func GenerateCpiJsonPayload(methodName string, rootTemplatePath string, replacementMap map[string]string) (string, error) {
	cpiTemplate := CpiTemplate{
		ID:             replacementMap["ID"],
		DirectorUuid:   replacementMap["DirectorUuid"],
		Tag_compiling:  replacementMap["Tag_compiling"],
		Tag_deployment: replacementMap["Tag_deployment"],
		Datacenter:     replacementMap["Datacenter"],
		VMID:           replacementMap["VMID"],
		DiskID:         replacementMap["DiskID"],
	}

	t := template.New(fmt.Sprintf("%s.json", methodName))

	methodPath := filepath.Join(rootTemplatePath, "/test_fixtures/cpi_methods", fmt.Sprintf("%s.json", methodName))
	t, err := t.ParseFiles(methodPath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, cpiTemplate)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func CreateTmpConfigPath(rootTemplatePath string, configPath string, username string, apiKey string) (string, error) {
	configTemplate := ConfigTemplate{
		Username: username,
		ApiKey:   apiKey,
	}

	t := template.New("config.json")

	t, err := t.ParseFiles(filepath.Join(rootTemplatePath, configPath))
	if err != nil {
		return "", err
	}

	tempFile, err := ioutil.TempFile("", "config")
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, configTemplate)
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}
