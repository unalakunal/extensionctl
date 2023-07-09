package image

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AppExtensionConfig struct {
	DockerfilePaths []string `json:"dockerfile_paths"`
	DirPath         string   `json:"dir_path"`
	KaapanaPath     string   `json:"kaapana_path"`
}

func ParseConfigFile(configPath string) (*AppExtensionConfig, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config AppExtensionConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func ValidateConfig(config *AppExtensionConfig) error {
	if config.DirPath == "" || config.KaapanaPath == "" {
		return errors.New("<dir_path> or <kaapana_path> is empty")
	}

	if !isAbsolutePath(config.DirPath) || !isAbsolutePath(config.KaapanaPath) {
		return errors.New("<dir_path> or <kaapana_path> is not a valid absolute path")
	}

	return nil
}

func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func GlobDockerfilePaths(config *AppExtensionConfig, configPath string) error {
	files, err := filepath.Glob(filepath.Join(config.DirPath, "/**/Dockerfile*"))
	if err != nil {
		fmt.Printf("error with Glob %s: %s", filepath.Join(config.DirPath, "/**/Dockerfile*"), err)
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no dockerfiles found in %s", filepath.Join(config.DirPath, "/**/Dockerfile*"))
	}

	config.DockerfilePaths = files

	return writeConfigFile(config, configPath)
}

func writeConfigFile(config *AppExtensionConfig, configPath string) error {
	file, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, file, 0644)
	if err != nil {
		return err
	}

	return nil
}

func FindPrereqDockerfiles(config *AppExtensionConfig) ([]string, error) {
	prereqDockerfiles := make([]string, 0)

	for _, dockerfile := range config.DockerfilePaths {
		lines, err := readLines(dockerfile)
		if err != nil {
			return nil, err
		}

		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "FROM local-only/") {
			fmt.Println(firstLine)
			imageName, _, err := getImageNameAndTagFromFirstLine(firstLine)
			if err != nil {
				fmt.Printf("Errorrr: %s", err)
				return nil, err
			}
			fmt.Println(imageName)
			dockerfilePaths, err := findDockerfilesInKaapanaPath(imageName, config.KaapanaPath)
			if err != nil {
				return nil, err
			}

			for _, dockerfilePath := range dockerfilePaths {
				prereqDockerfiles = appendIfUnique(prereqDockerfiles, dockerfilePath)
			}
		}
	}

	for {
		newPrereqDockerfiles := make([]string, 0)

		for _, prereqDockerfile := range prereqDockerfiles {
			lines, err := readLines(prereqDockerfile)
			if err != nil {
				return nil, err
			}

			firstLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(firstLine, "FROM local-only/") {
				fmt.Printf("firstLine: %s\n", firstLine)
				imageName := getImageNameFromLabel(firstLine)
				fmt.Printf("imageName: %s\n", imageName)
				dockerfilePaths, err := findDockerfilesInKaapanaPath(imageName, config.KaapanaPath)
				if err != nil {
					return nil, err
				}

				for _, dockerfilePath := range dockerfilePaths {
					if !contains(prereqDockerfiles, dockerfilePath) && !contains(newPrereqDockerfiles, dockerfilePath) {
						newPrereqDockerfiles = append(newPrereqDockerfiles, dockerfilePath)
					}
				}
			}
		}

		if len(newPrereqDockerfiles) == 0 {
			break
		}

		prereqDockerfiles = append(prereqDockerfiles, newPrereqDockerfiles...)
	}

	return prereqDockerfiles, nil
}

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

func getImageNameFromLabel(line string) string {
	split := strings.Split(line, "=")
	if len(split) == 2 {
		return strings.Trim(split[1], "\"'")
	}
	return ""
}

func getImageNameAndTagFromFirstLine(line string) (string, string, error) {
	parts := strings.Fields(line)

	if len(parts) < 2 {
		return "", "", errors.New("invalid first line format")
	}

	fromParts := strings.Split(parts[1], "/")

	if len(fromParts) < 2 {
		return "", "", errors.New("invalid first line format")
	}

	imageParts := strings.Split(fromParts[1], ":")

	if len(imageParts) < 2 {
		return "", "", errors.New("invalid first line format")
	}

	imageName := imageParts[0]
	imageTag := imageParts[1]

	return imageName, imageTag, nil
}

func findDockerfilesInKaapanaPath(imageName string, kaapanaPath string) ([]string, error) {
	fmt.Println("in f")
	var dockerfilePaths []string

	add_count := 0

	err := filepath.Walk(kaapanaPath, func(filePath string, info os.FileInfo, err error) error {
		// go through all the Dockerfiles inside kaapanaPath
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "Dockerfile" {
			lines, err := readLines(filePath)
			if err != nil {
				return err
			}

			for _, l := range lines {
				if strings.Contains(l, fmt.Sprintf("LABEL IMAGE=\"%s\"", imageName)) {
					dockerfilePaths = append(dockerfilePaths, filePath)
					add_count += 1
				}
			}
		}

		return nil
	})
	fmt.Printf("add_count for %s is %d\n", imageName, add_count)

	if err != nil {
		return nil, err
	}

	if len(dockerfilePaths) == 0 {
		return nil, errors.New("Dockerfile not found for image: " + imageName)
	}

	return dockerfilePaths, nil
}

func appendIfUnique(slice []string, element string) []string {
	for _, s := range slice {
		if s == element {
			return slice
		}
	}
	return append(slice, element)
}

func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}

func getLabelofDockerfile(dockerfile string) (string, error) {
	fileLines, err := readLines(dockerfile)
	if err != nil {
		return "", err
	}
	var labelLine string
	count := 0
	for _, s := range fileLines {
		if strings.Contains(s, "LABEL IMAGE=\"") {
			labelLine = s
			count += 1
		}
	}
	if count == 0 {
		return "", errors.New("failed to find line with LABEL IMAGE=\" in dockerfile " + dockerfile)
	} else if count > 1 {
		return "", errors.New("found multiple lines with LABEL IMAGE=\" in dockerfile " + dockerfile)
	}

	res := getImageNameFromLabel(labelLine)

	return res, nil
}

func BuildDockerImage(dockerfile string, prefix string) (string, error) {
	fmt.Printf("building docker image: %s\n", dockerfile)
	imageName, err := getLabelofDockerfile(dockerfile)
	if err != nil {
		return "", err
	}
	ctxPath := dockerfile
	suffix := "/Dockerfile"
	if strings.HasSuffix(ctxPath, suffix) {
		ctxPath, _ = strings.CutSuffix(ctxPath, suffix)
	}
	fmt.Printf("imageName %s, tag %s\n", imageName, prefix+imageName+":latest")
	command := exec.Command("docker", "build", "-t", prefix+imageName+":latest", ctxPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return "", errors.New("failed to build Docker image: " + err.Error())
	}

	fmt.Printf("successfully built %s in path %s\n", "local-only/"+imageName+":latest", dockerfile)

	return imageName, nil
}

func getFirstLine(filepath string) string {
	lines, err := readLines(filepath)
	if err != nil {
		return ""
	}
	return lines[0]
}

func BuildAndSaveImage(dirPath string, dockerfile string) error {
	imageName, err := BuildDockerImage(dockerfile, "docker.io/kaapana/")
	if err != nil {
		return err
	}

	savePath := filepath.Join(dirPath, imageName+".tar")
	fmt.Printf("saving image %s into %s...\n", "docker.io/kaapana/"+imageName, savePath)
	command := exec.Command("docker", "save", "docker.io/kaapana/"+imageName+":latest", "-o", savePath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return errors.New("failed to save Docker image: " + err.Error())
	}

	return nil
}
