VERSION := 0.2.11
LDFLAGS = "-X 'main.version=$(VERSION)'"
APP_NAME = terraform-provider-mongodb
OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH ?= arm64
BIN_DIR ?= $(shell go env GOPATH)/bin
EXAMPLES_DIR ?= ./examples
SCHEMA_FILE := providers-schema-short.json

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

# Generate provider schema from local build
generate-schema: build install
	@echo "Generating provider schema..."
	@mkdir -p $(TF_TEST_DIR)
	@cd $(TF_TEST_DIR) && \
		env TF_CLI_CONFIG_FILE=.terraformrc terraform providers schema -json > schema.json 2>&1 || true
	@if [ -s $(TF_TEST_DIR)/schema.json ]; then \
		cp $(TF_TEST_DIR)/schema.json providers-schema.json; \
		cat providers-schema.json | \
			jq '.provider_schemas.mongodb = .provider_schemas["registry.terraform.io/ahton89/mongodb"] | del(.provider_schemas["registry.terraform.io/ahton89/mongodb"])' \
			> $(SCHEMA_FILE); \
		echo "Generating provider schema... DONE"; \
	else \
		echo "ERROR: Failed to generate schema"; \
		exit 1; \
	fi

# Generate documentation using pre-generated schema
generate-docs: generate-schema
	@echo "Updating examples version..."
	@find $(EXAMPLES_DIR) -name "*.tf" -print0 | xargs -0 sed -i '' -E "s/(version = \")= [0-9]+\.[0-9]+\.[0-9]+(\")/\1= $(VERSION)\2/"
	@echo "Updating examples version... DONE"
	@echo "Generating documentation..."
	@tfplugindocs generate \
		--providers-schema $(SCHEMA_FILE) \
		--provider-name mongodb \
		--rendered-provider-name MongoDB
	@echo "Generating documentation... DONE"

# Validate documentation
validate-docs: generate-schema
	@echo "Validating documentation..."
	@tfplugindocs validate \
		--providers-schema $(SCHEMA_FILE) \
		--provider-name mongodb
	@echo "Validating documentation... DONE"

# Complete documentation workflow: generate + validate
docs: generate-schema generate-docs validate-docs
	@echo "Documentation generation and validation complete!"

.PHONY: build build-all install clean
.PHONY: generate-schema generate-docs validate-docs docs
