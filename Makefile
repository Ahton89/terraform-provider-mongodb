VERSION := 0.1.0
MONGODB_MIN_VERSION = 4.4.0
LDFLAGS = "-X 'main.version=$(VERSION)' -X 'terraform-provider-mongodb/internal/provider.minVersion=$(MONGODB_MIN_VERSION)'"
APP_NAME = terraform-provider-mongodb
OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH ?= arm64
BIN_DIR ?= $(shell go env GOPATH)/bin

build:
	@echo "Compiling $(APP_NAME) for $(OS)/$(ARCH) platform with version $(VERSION)..."
	@GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags=$(LDFLAGS) -o ./$(APP_NAME)-$(OS)-$(ARCH) ./main.go
	@echo "Compiling $(APP_NAME) for $(OS)/$(ARCH) platform with version $(VERSION)... DONE"

build-all:
	@$(MAKE) OS=darwin ARCH=amd64 build
	@$(MAKE) OS=darwin ARCH=arm64 build
	@$(MAKE) OS=windows ARCH=amd64 build
	@$(MAKE) OS=linux ARCH=amd64 build

install: build
	@echo "Installing $(APP_NAME) to $(BIN_DIR)..."
	@mkdir -p $(BIN_DIR)
	@mv ./$(APP_NAME)-$(OS)-$(ARCH) $(BIN_DIR)/$(APP_NAME)
	@echo "Installing $(APP_NAME) to $(BIN_DIR)... DONE"

clean:
	@echo "Cleaning up..."
	@rm -f ./$(APP_NAME)-*
	@echo "Cleaning up... DONE"

.PHONY: build build-all install clean