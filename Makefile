BINARY_NAME := extensionctl
BUILD_DIR := ./build
SRC_DIR := ./cmd
VERSION := 0.0.1

.DEFAULT_GOAL := build

.PHONY: build clean

build: clean
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/cmd.go $(SRC_DIR)/utils.go

clean:
	rm -rf $(BUILD_DIR)

release: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(SRC_DIR)/cmd.go $(SRC_DIR)/utils.go
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(SRC_DIR)/cmd.go $(SRC_DIR)/utils.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(SRC_DIR)/cmd.go $(SRC_DIR)/utils.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(SRC_DIR)/cmd.go $(SRC_DIR)/utils.go

