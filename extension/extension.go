package extension

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Extension struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Description      string   `json:"description"`
	HelmStatus       string   `json:"helm_status"`
	KubernetesStatus []string `json:"kubernetes_status"`
}

func GetExtensions() ([]Extension, error) {
	chartsDir := "/home/kaapana/extensions" // Update this with the actual path to your charts directory

	files, err := os.ReadDir(chartsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory: %w", err)
	}

	extensions := make([]Extension, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext, err := extractExtensionInfo(filepath.Join(chartsDir, file.Name()))
		if err != nil {
			fmt.Printf("Error extracting extension info: %v", err)
			continue
		}

		helmStatus, err := getHelmStatus(ext.Name)
		if err != nil {
			fmt.Printf("Error getting Helm status for %s: %v", ext.Name, err)
		}
		ext.HelmStatus = helmStatus

		kubernetesStatus, err := getKubernetesStatus(ext.Name)
		if err != nil {
			fmt.Printf("Error getting Kubernetes status: %v", err)
		}
		ext.KubernetesStatus = kubernetesStatus

		extensions = append(extensions, ext)
	}

	return extensions, nil
}

func extractExtensionInfo(filePath string) (Extension, error) {
	ext := Extension{}

	file, err := os.Open(filePath)
	if err != nil {
		return ext, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return ext, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)

	foundChartYAML := false
	foundValuesYAML := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ext, fmt.Errorf("failed to read tar header: %w", err)
		}

		fileName := path.Base(header.Name)

		switch fileName {
		case "Chart.yaml":
			// Read and parse the Chart.yaml file to extract extension information
			chartData, err := extractFileContents(tarReader)
			if err != nil {
				return ext, fmt.Errorf("failed to extract Chart.yaml contents: %w", err)
			}

			var chartMetadata struct {
				Name        string `yaml:"name"`
				Version     string `yaml:"version"`
				Description string `yaml:"description"`
			}
			err = yaml.Unmarshal([]byte(chartData), &chartMetadata)
			if err != nil {
				return ext, fmt.Errorf("failed to unmarshal Chart.yaml: %w", err)
			}

			ext.Name = chartMetadata.Name
			ext.Version = chartMetadata.Version
			ext.Description = chartMetadata.Description

			foundChartYAML = true

		case "values.yaml":
			// Read and parse the values.yaml file to extract additional extension information
			valuesData, err := extractFileContents(tarReader)
			if err != nil {
				return ext, fmt.Errorf("failed to extract values.yaml contents: %w", err)
			}

			_, err = parseValuesYAML(valuesData)
			if err != nil {
				return ext, fmt.Errorf("failed to parse values.yaml: %w", err)
			}

			// // Extract additional extension information from the values map
			// // Example: Extracting the multiinstallable flag from the values
			// if multiinstallable, ok := values["multiinstallable"]; ok {
			// 	ext.Multiinstallable = multiinstallable.(bool)
			// }

			foundValuesYAML = true
		}

		// Break the loop if we have found both Chart.yaml and values.yaml
		if foundChartYAML && foundValuesYAML {
			break
		}
	}

	return ext, nil
}

func extractFileContents(tarReader *tar.Reader) (string, error) {
	data, err := io.ReadAll(tarReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file contents: %w", err)
	}
	return string(data), nil
}

func parseValuesYAML(valuesData string) (interface{}, error) {
	var values interface{}
	err := yaml.Unmarshal([]byte(valuesData), &values)
	if err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	return values, nil
}

func getHelmStatus(extensionName string) (string, error) {
	// Run helm status command and capture the output
	cmd := exec.Command("helm", "status", extensionName)
	output, err := cmd.Output()
	if err != nil {
		// Check if the command returned a non-zero exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("helm status command failed with exit code %d: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return "", fmt.Errorf("failed to execute helm status command: %w", err)
	}

	// Extract the Helm status from the output
	status := string(output)

	return status, nil
}

func getKubernetesStatus(extensionName string) ([]string, error) {
	// Run kubectl get command and capture the output
	cmd := exec.Command("microk8s.kubectl", "get", "all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl get command: %w", err)
	}

	// Process the output and extract the resource statuses
	statuses := parseKubernetesStatus(string(output))

	return statuses, nil
}

func parseKubernetesStatus(output string) []string {
	lines := strings.Split(output, "\n")
	statuses := make([]string, 0)

	// Iterate over the lines and extract the resource statuses
	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line by whitespace
		fields := strings.Fields(line)

		// The status is typically in the third column
		if len(fields) >= 3 {
			status := fields[2]
			statuses = append(statuses, status)
		}
	}

	return statuses
}
