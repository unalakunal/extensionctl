package util

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

type ExtensionConfig struct {
	DockerfilePaths      []string `json:"dockerfile_paths"`
	DirPath              string   `json:"dir_path"`
	KaapanaPath          string   `json:"kaapana_path"`
	KaapanaBuildVersion  string   `json:"kaapana_build_version"`
	NoSave               bool     `json:"no_save"`
	NoRebuild            bool     `json:"no_rebuild"`
	NoOverwriteOperators bool     `json:"no_overwrite_operators"`
	CustomRegistryUrl    string   `json:"custom_registry_url"`
	ContainerEngine      string   `json:"container_engine"`
	ChartPath            string   `json:"chart_path"`
}

func ParseConfigFile(configPath string, noSave bool, noRebuild bool) (*ExtensionConfig, error) {
	color.Blue("parsing config file")
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ExtensionConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	config.NoSave = noSave
	config.NoRebuild = noRebuild

	if config.DockerfilePaths == nil {
		config.DockerfilePaths = []string{}
	}

	if config.KaapanaBuildVersion == "" || config.CustomRegistryUrl == "" {
		deployment, err := KubeGetDeployment("kube-helm-deployment", "admin")
		if err != nil {
			return nil, err
		}
		if config.KaapanaBuildVersion == "" {
			version, err := GetEnvVarFromDeployment(deployment, "KAAPANA_BUILD_VERSION")
			if err != nil {
				return nil, err
			}
			color.Magenta("setting KaapanaBuildVersion in config as %s", version)
			config.KaapanaBuildVersion = version
		}
		if config.CustomRegistryUrl == "" {
			registryURL, err := GetEnvVarFromDeployment(deployment, "REGISTRY_URL")
			if err != nil {
				return nil, err
			}
			color.Magenta("setting CustomRegistryUrl in config as %s", registryURL)
			config.CustomRegistryUrl = registryURL
		}
	}

	err = WriteConfigFile(&config, configPath)
	if err != nil {
		return nil, err
	}

	// TODO: make sure CustomRegistryUrl doesn't start with "https://" and doesn't end with "/"

	return &config, nil
}

func WriteConfigFile(config *ExtensionConfig, configPath string) error {
	file, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		color.Red(err.Error())
		return err
	}

	err = os.WriteFile(configPath, file, 0644)
	if err != nil {
		color.Red(err.Error())
		return err
	}

	return nil
}

func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func ValidateConfig(dirPath string, kaapanaPath string) error {
	if dirPath == "" || kaapanaPath == "" {
		err := errors.New("<dir_path> or <kaapana_path> is empty")
		color.Red(err.Error())
		return err
	}

	if !isAbsolutePath(dirPath) || !isAbsolutePath(kaapanaPath) {
		err := errors.New("<dir_path> or <kaapana_path> is not a valid absolute path")
		color.Red(err.Error())
		return err
	}

	return nil
}
