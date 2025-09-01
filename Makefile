VERSION := 0.2.8
LDFLAGS = "-X 'main.version=$(VERSION)'"
APP_NAME = terraform-provider-mongodb
OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH ?= arm64
BIN_DIR ?= $(shell go env GOPATH)/bin
EXAMPLES_DIR ?= ./_examples

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

preparing-examples:
	@echo "Updating examples version..."
	@find $(EXAMPLES_DIR) -name "*.tf" -print0 | xargs -0 sed -i '' -E "s/(version = \")= [0-9]+\.[0-9]+\.[0-9]+(\")/\1= $(VERSION)\2/"
	@echo "Updating examples version... DONE"

preparing-docs:
	@echo "Generating docs..."
	@tfplugindocs generate --examples-dir $(EXAMPLES_DIR)
	@echo "Generating docs... DONE"

preparing: preparing-examples preparing-docs

.PHONY: build build-all install clean preparing-examples preparing-docs preparing
