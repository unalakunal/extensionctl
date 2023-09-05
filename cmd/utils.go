package main

import (
	"errors"
	"path/filepath"

	"github.com/fatih/color"
)

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
