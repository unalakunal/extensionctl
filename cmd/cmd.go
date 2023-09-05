package main

import (
	"extensionctl/chart"
	"extensionctl/extension"
	"extensionctl/image"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func ImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image [json file]",
		Short: "Build and save Docker images",
		Args:  cobra.ExactArgs(1),
		RunE:  buildImages,
	}

	return cmd
}

func ChartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart [json file]",
		Short: "Package the Helm chart",
		Args:  cobra.ExactArgs(1),
		RunE:  packageChart,
	}

	return cmd
}

func packageChart(cmd *cobra.Command, args []string) error {
	noColor, _ := cmd.Flags().GetBool("no_color")
	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}

	color.Magenta("Packaging helm chart...")
	configPath := args[0]
	config, err := chart.ParseConfigFile(configPath)
	if err != nil {
		return err
	}
	fmt.Println("parsed config file")

	if err := ValidateConfig(config.DirPath, config.KaapanaPath); err != nil {
		return err
	}
	fmt.Println("config validated")

	// check requirements yaml
	// if exists, helm dep up
	// check if deployment.yaml or job.yaml exists under /charts/templates or /charts/<chart-name>/templates
	resourceYaml, err := chart.FindResourceYaml(config)
	valuesYaml, err := chart.FindValuesYaml(config)

	// add to values.yaml -> custom_registry_url: "docker.io/kaapana"
	// add to values.yaml -> pull_policy_images: "IfNotPresent"
	// add to values.yaml ->
	err = chart.EditValuesYaml(valuesYaml, config)

	// TODO: probably not necessary at all
	err = chart.EditResourceYaml(resourceYaml, config)

	helmChart, err := chart.PackageChart(config)

	color.Blue("Successfully packaged Helm chart as %s", helmChart)

	return nil
}

func buildAll(cmd *cobra.Command, args []string) error {
	noColor, _ := cmd.Flags().GetBool("no_color")
	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}
	color.Magenta("building both charts and images")
	return nil
}

func getExtensions(cmd *cobra.Command, args []string) error {
	noColor, _ := cmd.Flags().GetBool("no_color")
	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}
	extensions, err := extension.GetExtensions()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Print the extensions
	for _, ext := range extensions {
		fmt.Printf("Name: %s\n", ext.Name)
		fmt.Printf("Version: %s\n", ext.Version)
		fmt.Printf("Description: %s\n", ext.Description)
		fmt.Printf("Helm Status: %s\n", ext.HelmStatus)
		fmt.Printf("Kubernetes Status:\n")
		for _, status := range ext.KubernetesStatus {
			fmt.Printf("- %s\n", status)
		}
		fmt.Println("--------")
	}
	return nil
}

func buildImages(cmd *cobra.Command, args []string) error {
	noColor, _ := cmd.Flags().GetBool("no_color")
	noSave, _ := cmd.Flags().GetBool("no_save")
	noRebuild, _ := cmd.Flags().GetBool("no_rebuild")

	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}
	color.Magenta("Building images...")
	configPath := args[0]
	config, err := image.ParseConfigFile(configPath, noSave, noRebuild)
	if err != nil {
		return err
	}
	fmt.Println("parsed config file")

	if err := ValidateConfig(config.DirPath, config.KaapanaPath); err != nil {
		return err
	}
	fmt.Println("config validated")

	if len(config.DockerfilePaths) == 0 {
		if err := image.GlobDockerfilePaths(config, configPath); err != nil {
			return err
		}
	}
	fmt.Printf("config.DockerfilePaths: %s\n", config.DockerfilePaths)

	prereqDockerfiles, err := image.FindPrereqDockerfiles(config)
	if err != nil {
		return err
	}

	// change image references
	if config.NoOverwriteOperators != true {
		if err := image.ChangeImageRefs(config.DirPath, "{DEFAULT_REGISTRY}", config.CustomRegistryUrl); err != nil {
			return err
		}
		if err := image.ChangeImageRefs(config.DirPath, "{KAAPANA_BUILD_VERSION}\",", config.KaapanaBuildVersion+"\",\nimage_pull_policy=\"IfNotPresent\",\n"); err != nil {
			return err
		}
	} else {
		color.Blue("skipping changing operator py files")
	}

	for _, prereqDockerfile := range prereqDockerfiles {
		if _, err := image.BuildDockerImage(prereqDockerfile, config, true); err != nil {
			return err
		}
	}

	imageTags := []string{}
	for _, dockerfile := range config.DockerfilePaths {
		imageTag, err := image.BuildDockerImage(dockerfile, config, false)
		if err != nil {
			color.Red("Failed to build image: %s , err: %s", imageTags, err.Error())
			return err
		}
		imageTags = append(imageTags, imageTag)
	}
	err = image.SaveImages(imageTags, config.DirPath)
	if err != nil {
		color.Red("Failed to save images: %s , err: %s", imageTags, err.Error())
		return err
	}

	color.Blue("Successfully built and saved the images.")
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "extensionctl",
		Short: "Extension Manager CLI",
		Long:  "A command-line tool for managing extensions",
		Run: func(cmd *cobra.Command, args []string) {
			// Display help information if no command is specified
			cmd.Help()
		},
	}

	extensionsCmd := &cobra.Command{
		Use:   "extensions",
		Short: "Get extensions",
		Long:  "Get all extensions",
		RunE:  getExtensions,
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "generate chart tgz and build/save Docker images",
		Args:  cobra.ExactArgs(1),
		RunE:  buildAll,
	}

	// Flags
	rootCmd.PersistentFlags().BoolP("no_color", "c", false, "Disable color")
	rootCmd.PersistentFlags().BoolP("no_save", "s", false, "Disable image save")
	rootCmd.PersistentFlags().BoolP("no_rebuild", "b", false, "Disable rebuilding existing images")
	rootCmd.PersistentFlags().BoolP("no_overwrite_operators", "w", false, "Disable searching and replacing patterns in py files")

	// Add subcommands for different functionalities
	rootCmd.AddCommand(extensionsCmd)
	rootCmd.AddCommand(buildCmd)
	buildCmd.AddCommand(ImageCmd())
	buildCmd.AddCommand(ChartCmd())

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {

		fmt.Println(err)
		os.Exit(1)
	}
}
