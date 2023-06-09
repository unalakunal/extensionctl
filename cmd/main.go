package main

import (
	"extension-manager/extensions"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "extension-manager",
		Short: "Extension Manager CLI",
		Long:  "A command-line tool for managing extensions",
		Run: func(cmd *cobra.Command, args []string) {
			// Display help information if no command is specified
			cmd.Help()
		},
	}

	// Add subcommands for different functionalities
	rootCmd.AddCommand(getExtensionsCmd())

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
			extensions, err := extensions.GetExtensions()
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
