package main

import (
	"extensionctl/chart"
	"extensionctl/extension"
	"extensionctl/image"
	"extensionctl/util"
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
	noSave, _ := cmd.Flags().GetBool("no_save")
	noRebuild, _ := cmd.Flags().GetBool("no_rebuild")
	noColor, _ := cmd.Flags().GetBool("no_color")

	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}

	color.Magenta("Building images...")
	if noColor {
		os.Setenv("NO_COLOR", "TRUE")
	}

	color.Magenta("Packaging helm chart...")
	configPath := args[0]
	config, err := util.ParseConfigFile(configPath, noSave, noRebuild)
	if err != nil {
		color.Red("failed to Parse config file %s", err.Error())
		return err
	}
	color.White("parsed config file")

	if err := util.ValidateConfig(config.DirPath, config.KaapanaPath); err != nil {
		return err
	}
	color.White("config validated")

	// chart path
	config, err = chart.FindChartPath(config)
	if err != nil {
		color.Red("failed to find the chart folder containing Chart.yaml under %s", config.DirPath)
		return err
	}
	err = util.WriteConfigFile(config, configPath)
	if err != nil {
		return err
	}
	color.Magenta("ChartPath is set as %s", config.ChartPath)

	// requirements
	err = chart.HandleRequirements(config)
	if err != nil {
		color.Red("failed to update requirements %s", err.Error())
		return err
	}
	color.Magenta("Succesfully updated chart requirements")

	// Chart.yaml
	chart.EditChartYaml(config)
	if err != nil {
		color.Red("failed to update Chart.yaml %s", err.Error())
		return err
	}
	color.Magenta("Succesfully updated Chart.yaml")

	// values.yaml
	err = chart.EditValuesYaml(config)
	if err != nil {
		color.Red("failed to update values.yaml %s", err.Error())
		return err
	}
	color.Magenta("Succesfully updated values.yaml")

	// package
	err = chart.PackageChart(config)
	if err != nil {
		color.Red("failed to package chart %s", err.Error())
		return err
	}

	color.Magenta("Successfully packaged Helm chart")

	return nil
}

func buildAll(cmd *cobra.Command, args []string) error {
	color.Green("Building images and packaging charts")
	err := buildImages(cmd, args)
	if err != nil {
		return err
	}
	err = packageChart(cmd, args)
	if err != nil {
		return err
	}
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
	config, err := util.ParseConfigFile(configPath, noSave, noRebuild)
	if err != nil {
		return err
	}

	if err := util.ValidateConfig(config.DirPath, config.KaapanaPath); err != nil {
		return err
	}
	color.Blue("config validated")

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
	if !config.NoOverwriteOperators {
		if err := image.ChangeImageRefs(config.DirPath, "{DEFAULT_REGISTRY}", config.CustomRegistryUrl); err != nil {
			return err
		}
		if err := image.ChangeImageRefs(config.DirPath, "{KAAPANA_BUILD_VERSION}\",", config.KaapanaBuildVersion+"\",\nimage_pull_policy=\"IfNotPresent\",\n"); err != nil {
			return err
		}
	} else {
		color.Blue("skipping changing operator py files")
	}

	prereqDockerfiles, err = image.PrioritizePrereqs(prereqDockerfiles)
	if err != nil {
		return err
	}
	color.Blue("prioritized prereqDockerfiles %s", prereqDockerfiles)

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
	err = image.SaveImages(imageTags, config.DirPath, config.ContainerEngine)
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
		Short: "generate chart tgz, build and save Docker images",
		Args:  cobra.ExactArgs(1),
		RunE:  buildAll,
	}

	// Flags
	rootCmd.PersistentFlags().BoolP("no_color", "c", false, "disable colored terminal output")
	rootCmd.PersistentFlags().BoolP("no_save", "s", false, "disable saving images as .tar files")
	rootCmd.PersistentFlags().BoolP("no_rebuild", "b", false, "disable rebuilding existing images")
	rootCmd.PersistentFlags().BoolP("no_overwrite_operators", "w", false, "disable searching and replacing patterns in py files")

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
