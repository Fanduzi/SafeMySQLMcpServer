.PHONY: build run test clean install lint

# Binary names
BINARY_NAME=safe-mysql-mcp
TOKEN_TOOL=mysql-mcp-token

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main build targets
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	$(GOBUILD) -o $(BUILD_DIR)/$(TOKEN_TOOL) ./pkg/token

run:
	$(GOCMD) run ./cmd/server -config config/config.yaml

test:
	$(GOTEST) -v -race -cover ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

install-deps:
	$(GOMOD) download
	$(GOMOD) tidy

lint:
	golangci-lint run ./...

# Development helpers
token:
	@echo "Usage: ./bin/$(TOKEN_TOOL) --user <user_id> --email <email> --secret <jwt_secret>"
	@echo "Example: ./bin/$(TOKEN_TOOL) --user zhangsan --email zhangsan@company.com --expire 365d"

dev:
	$(GOCMD) run ./cmd/server -config config/config.yaml

# Docker (optional)
docker-build:
	docker build -t safe-mysql-mcp:latest .

docker-run:
	docker run -p 8080:8080 -v $(PWD)/config:/app/config safe-mysql-mcp:latest
