package chart

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type ExtensionConfig struct {
	DockerfilePaths     []string `json:"dockerfile_paths"`
	DirPath             string   `json:"dir_path"`
	KaapanaPath         string   `json:"kaapana_path"`
	KaapanaBuildVersion string   `json:"kaapana_build_version"`
	CustomRegistryUrl   string   `json:"custom_registry_url"`
}

func yamlRead(yamlPath string) (map[interface{}]interface{}, error) {
	color.Blue("reading from yaml file %s", yamlPath)

	f, err := os.ReadFile(yamlPath)
	if err != nil {
		color.Red(err.Error())
		return nil, err
	}

	data := make(map[interface{}]interface{})

	err = yaml.Unmarshal(f, &data)

	if err != nil {
		color.Red(err.Error())
		return nil, err
	}

	color.Blue("successfully read from %s", yamlPath)

	return data, nil
}

func yamlWrite(yamlPath string, data map[interface{}]interface{}) error {
	color.Blue("writing to yaml file %s", yamlPath)
	f, err := yaml.Marshal(data)

	if err != nil {
		color.Red(err.Error())
		return err
	}

	err = os.WriteFile(yamlPath, f, 0)

	if err != nil {
		color.Red(err.Error())
		return err
	}

	color.Blue("successfully written to %s", yamlPath)

	return nil
}

func ParseConfigFile(configPath string) (*ExtensionConfig, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ExtensionConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func ValidateConfig(config *ExtensionConfig) error {
	if config.DirPath == "" || config.KaapanaPath == "" {
		err := errors.New("<dir_path> or <kaapana_path> is empty")
		color.Red(err.Error())
		return err
	}

	if !isAbsolutePath(config.DirPath) || !isAbsolutePath(config.KaapanaPath) {
		err := errors.New("<dir_path> or <kaapana_path> is not a valid absolute path")
		color.Red(err.Error())
		return err
	}

	return nil
}

func FindResourceYaml(config *ExtensionConfig) (string, error) {
	return "", nil
}

func FindValuesYaml(config *ExtensionConfig) (string, error) {
	return "", nil
}

func EditValuesYaml(valuesYaml string, config *ExtensionConfig) error {
	return nil
}

func EditResourceYaml(resourceYaml string, config *ExtensionConfig) error {
	return nil
}

func AddValuesYaml(config *ExtensionConfig) error {
	return nil
}

func PackageChart(config *ExtensionConfig) (string, error) {
	return "", nil
}
