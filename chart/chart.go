package chart

import (
	"bufio"
	"errors"
	"extensionctl/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

func readLines(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
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

func findYamlInChartPath(chartPath string, fileName string) (string, error) {
	foundPath := ""

	err := filepath.Walk(chartPath, func(filePath string, info os.FileInfo, err error) error {
		// go through the chartPath to find yaml files
		if err != nil {
			color.Red(err.Error())
			return err
		}

		if !info.IsDir() && info.Name() == fileName {
			color.Blue("found yaml file '%s' under %s", fileName, filePath)
			foundPath = filePath
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	return foundPath, nil
}

func FindChartPath(config *util.ExtensionConfig) (*util.ExtensionConfig, error) {
	foundChart := ""

	err := filepath.Walk(config.DirPath, func(filePath string, info os.FileInfo, err error) error {
		// go through all the Dockerfiles inside kaapanaPath
		if err != nil {
			color.Red(err.Error())
			return err
		}

		if !info.IsDir() && info.Name() == "Chart.yaml" {
			if strings.Contains(filePath, "/charts/") {
				color.Yellow("found sub-chart Chart.yaml file %s , skipping ", filePath)
			} else {
				color.Blue("found Chart.yaml in path %s", filePath)
				foundChart = filePath
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	config.ChartPath, _ = strings.CutSuffix(foundChart, "/Chart.yaml")
	return config, nil
}

func HandleRequirements(config *util.ExtensionConfig) error {
	// check if requirements.yaml exists
	reqYaml, err := findYamlInChartPath(config.ChartPath, "requirements.yaml")
	if err != nil {
		return err
	}
	if reqYaml == "" {
		color.Magenta("No 'requirements.yaml' found under %s , skipped running 'helm dep up' this step", config.ChartPath)
		return nil
	}
	// if exists, helm dep up and untar in place
	color.Blue("running helm dep up %s --debug", config.ChartPath)
	command := exec.Command("helm", "dep", "up", config.ChartPath, "--debug")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return errors.New("failed to run 'helm dep up " + reqYaml + "': " + err.Error())
	}

	return nil
}

func EditChartYaml(config *util.ExtensionConfig) error {
	// Changes 'version' to config.KaapanaBuildVersion

	// read file
	chartFile := config.ChartPath + "/Chart.yaml"
	chartYaml, err := yamlRead(chartFile)
	if err != nil {
		color.Red("failed to read Chart.yaml %s", chartFile)
		return err
	}

	// change version
	_, ok := chartYaml["version"]
	if !ok {
		color.Red("Chart.yaml must contain a 'version' key %s", chartFile)
		return errors.New(chartFile + " does not have a 'version key")
	}
	chartYaml["version"] = config.KaapanaBuildVersion

	// write back
	err = yamlWrite(chartFile, chartYaml)
	if err != nil {
		color.Red("failed to write to Chart.yaml %s", chartFile)
		return err
	}
	return nil
}

func EditValuesYaml(config *util.ExtensionConfig) error {
	/* Adds
	 * custom_registry_url: config.CustomRegistryUrl
	 * pull_policy_images: IfNotPresent
	 */

	// read file
	valuesFile := config.ChartPath + "/values.yaml"
	valuesYaml, err := yamlRead(valuesFile)
	if err != nil {
		color.Red("failed to read Chart.yaml %s", valuesFile)
		return err
	}

	// add keys & values
	global := valuesYaml["global"].(map[string]interface{})
	global["custom_registry_url"] = config.CustomRegistryUrl
	global["pull_policy_images"] = "IfNotPresent"
	valuesYaml["global"] = global

	// write back
	err = yamlWrite(valuesFile, valuesYaml)
	if err != nil {
		color.Red("failed to write to values.yaml %s", valuesFile)
		return err
	}
	return nil
}

func PackageChart(config *util.ExtensionConfig) error {
	color.Blue("running helm package %s -d %s --debug", config.ChartPath, config.ChartPath)
	command := exec.Command("helm", "package", config.ChartPath, "-d", config.ChartPath, "--debug")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return errors.New("failed to run 'helm package " + config.ChartPath + "': " + err.Error())
	}

	return nil
}
