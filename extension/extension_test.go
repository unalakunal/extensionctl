package extension

import (
	"fmt"
	"testing"
)

func TestGetExtensions(t *testing.T) {
	extensions, err := GetExtensions()
	if err != nil {
		t.Fatalf("Failed to get extensions: %v", err)
	}

	fmt.Printf("Found %d extensions:\n", len(extensions))
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
}
