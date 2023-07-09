package main

import (
	"extensionctl/extension"
	"extensionctl/image"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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

	// Add subcommands for different functionalities
	rootCmd.AddCommand(getExtensionsCmd())
	rootCmd.AddCommand(buildImageCmd())

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getExtensionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "Get extensions",
		Long:  "Get all extensions",
		Run: func(cmd *cobra.Command, args []string) {
			// Call your GetExtensions function here
			extensions, err := extension.GetExtensions()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
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
		},
	}

	return cmd
}

func buildImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build and save Docker images",
		Args:  cobra.ExactArgs(1),
		RunE:  buildImages,
	}

	return cmd
}

func buildImages(cmd *cobra.Command, args []string) error {
	configPath := args[0]
	config, err := image.ParseConfigFile(configPath)
	if err != nil {
		return err
	}
	fmt.Println("parsed config file")

	if err := image.ValidateConfig(config); err != nil {
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

	for _, prereqDockerfile := range prereqDockerfiles {
		if _, err := image.BuildDockerImage(prereqDockerfile, "local-only/"); err != nil {
			return err
		}
	}
	fmt.Println("movin on")

	for _, dockerfile := range config.DockerfilePaths {
		if err := image.BuildAndSaveImage(config.DirPath, dockerfile); err != nil {
			return err
		}
	}

	fmt.Println("Successfully built and saved the images.")
	return nil
}
