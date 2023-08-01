package image

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type AppExtensionConfig struct {
	DockerfilePaths      []string `json:"dockerfile_paths"`
	DirPath              string   `json:"dir_path"`
	KaapanaPath          string   `json:"kaapana_path"`
	KaapanaBuildVersion  string   `json:"kaapana_build_version"`
	NoSave               bool     `json:"no_save"`
	NoRebuild            bool     `json:"no_rebuild"`
	NoOverwriteOperators bool     `json:"no_overwrite_operators"`
	CustomRegistryUrl    string   `json:"custom_registry_url"`
}

func ParseConfigFile(configPath string, noSave bool, noRebuild bool) (*AppExtensionConfig, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config AppExtensionConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	config.NoSave = noSave
	config.NoRebuild = noRebuild

	// TODO: make sure CustomRegistryUrl doesn't start with "https://" and doesn't end with "/"

	return &config, nil
}

func ValidateConfig(config *AppExtensionConfig) error {
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

func visitFile(path string, info os.DirEntry, err error) error {
	if err != nil {
		color.Red("Encountered error: %v\n", err)
		return err
	}

	if info.IsDir() {
		// Skip directories
		return nil
	}

	fmt.Println(path)
	return nil
}

func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func GlobDockerfilePaths(config *AppExtensionConfig, configPath string) error {
	var dockerfilePaths []string
	err := filepath.WalkDir(config.DirPath, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			color.Red("Encountered error: %s\n", err.Error())
			return nil
		}

		if info.IsDir() {
			// Skip directories
			return nil
		}

		if strings.HasSuffix(path, "Dockerfile") {
			fmt.Println(path)
			dockerfilePaths = append(dockerfilePaths, path)
		}
		return nil
	})
	fmt.Printf("dockerfilePaths: %s", dockerfilePaths)
	if err != nil {
		color.Red("Encountered error while walking directory: %s\n", err.Error())
	}

	config.DockerfilePaths = dockerfilePaths

	return writeConfigFile(config, configPath)
}

func writeConfigFile(config *AppExtensionConfig, configPath string) error {
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

func FindPrereqDockerfiles(config *AppExtensionConfig) ([]string, error) {
	prereqDockerfiles := make([]string, 0)

	for _, dockerfile := range config.DockerfilePaths {
		lines, err := readLines(dockerfile)
		if err != nil {
			color.Red(err.Error())
			return nil, err
		}

		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "FROM local-only/") {
			imageName, _, err := getImageNameAndTagFromFirstLine(firstLine)
			if err != nil {
				color.Red(err.Error())
				return nil, err
			}
			fmt.Printf("found local prerequisite image %s\n", imageName)
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
				imageName := getImageNameFromFirstLine(firstLine)
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

func getImageNameFromFirstLine(line string) string {
	fmt.Printf("getImageNameFromFirstLine %s\n", line)
	split := strings.Split(line, ":")

	if len(split) == 2 {
		trimmed := strings.Split(split[0], "/")
		return trimmed[len(trimmed)-1]
	}
	color.Red("split failed in getImageNameFromFirstLine %s", split)
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
	var dockerfilePaths []string

	add_count := 0

	err := filepath.Walk(kaapanaPath, func(filePath string, info os.FileInfo, err error) error {
		// go through all the Dockerfiles inside kaapanaPath
		if err != nil {
			color.Red(err.Error())
			return err
		}

		if !info.IsDir() && info.Name() == "Dockerfile" {
			lines, err := readLines(filePath)
			if err != nil {
				color.Red(err.Error())
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

func BuildDockerImage(dockerfile string, config *AppExtensionConfig, localOnly bool) (string, error) {
	color.Blue("building docker image: %s\n", dockerfile)
	imageName, err := getLabelofDockerfile(dockerfile)
	if err != nil {
		return "", err
	}
	ctxPath := dockerfile
	suffix := "/Dockerfile"
	if strings.HasSuffix(ctxPath, suffix) {
		ctxPath, _ = strings.CutSuffix(ctxPath, suffix)
	}
	registry := config.CustomRegistryUrl
	if localOnly {
		registry = "local-only"
	}
	tag := registry + "/" + imageName + ":" + config.KaapanaBuildVersion
	if imageExists(tag) && config.NoRebuild {
		color.Yellow("image %s already exists, not building since no_rebuild==true", tag)
		return imageName, nil
	}
	color.Blue("imageName %s, tag %s\n", imageName, tag)
	command := exec.Command("docker", "build", "-t", tag, ctxPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return "", errors.New("failed to build Docker image: " + err.Error())
	}

	color.Magenta("successfully built %s in path %s\n", tag, dockerfile)

	return tag, nil
}

func getFirstLine(filepath string) string {
	lines, err := readLines(filepath)
	if err != nil {
		return ""
	}
	return lines[0]
}

func SaveImages(imageNames []string, dirPath string) error {
	// save
	savePath := filepath.Join(dirPath, "images.tar")
	cmd := []string{"save", "-o", savePath}
	for _, imageName := range imageNames {
		cmd = append(cmd, imageName)
	}
	color.Blue("saving images %s into %s...\n", imageNames, savePath)
	command := exec.Command("docker", cmd...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		color.Red("failed to save Docker images: %s", err.Error())
		return errors.New("failed to save Docker images: " + err.Error())
	}

	return nil
}

func ChangeImageRefs(dirPath string, query string, newValue string) error {
	color.Blue("Changing image references in .py files")
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			color.Red("Error accessing path %s: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".py") {
			color.Magenta("file: %s , changing '%s' to '%s'", path, query, newValue)
			err := searchAndReplace(path, query, newValue)
			if err != nil {
				color.Red("Error searching and replacing in file %s: %v\n", path, err)
			}
			return nil
		}

		return nil
	})

	if err != nil {
		color.Red("Error walking through directory: %v\n", err)
		return err
	}
	return nil
}

func searchAndReplace(file string, query string, newValue string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		color.Red(err.Error())
		return err
	}

	if strings.Contains(string(content), query) {
		fmt.Printf("Changing %s to %s in file %s\n", query, newValue, file)
	}
	newContent := strings.ReplaceAll(string(content), query, newValue)
	err = ioutil.WriteFile(file, []byte(newContent), 0)
	if err != nil {
		color.Red(err.Error())
		return err
	}

	return nil
}

func imageExists(image string) bool {
	out, err := exec.Command("docker", "images", image).Output()
	if err != nil {
		color.Red(err.Error())
	}

	numLines := strings.Count(string(out), "\n")
	if numLines > 1 {
		return true
	}

	return false
}
