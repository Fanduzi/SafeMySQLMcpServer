.PHONY: build run test clean install lint \
       init dev deploy stop logs token \
       docker-build validate

# ── Project settings ──
BINARY_NAME=safe-mysql-mcp
TOKEN_TOOL=mysql-mcp-token
BUILD_DIR=bin

# ── Go commands ──
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

# ── Docker settings ──
DOCKER_COMPOSE=docker compose
DOCKER_COMPOSE_DEV=$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml

# ────────────────────────────────
# Setup & Deploy
# ────────────────────────────────

## init: Interactive setup — generates .env and config.yaml
init:
	@bash scripts/init.sh

## validate: Check required config files before deploy
validate:
	@test -f .env || { echo "Error: .env not found. Run 'make init' first."; exit 1; }
	@grep -q '^JWT_SECRET=.' .env || { echo "Error: JWT_SECRET is empty in .env. Run 'make init'."; exit 1; }
	@test -f config/config.yaml || { echo "Error: config/config.yaml not found. Run 'make init' first."; exit 1; }
	@grep -q 'host:' config/config.yaml || { echo "Error: config.yaml missing cluster host. Run 'make init'."; exit 1; }
	@echo "Validation passed."

## deploy: Build image and start production container
deploy: validate
	$(DOCKER_COMPOSE) up -d --build
	@echo ""
	@echo "Deployed. Status:"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "Generate a token: make token"

## dev: Start development environment (app + MySQL)
dev:
	$(DOCKER_COMPOSE_DEV) up -d --build
	@echo ""
	@echo "Dev environment started. Status:"
	@$(DOCKER_COMPOSE_DEV) ps

## stop: Stop all containers
stop:
	$(DOCKER_COMPOSE) down 2>/dev/null; true

## logs: Follow app logs
logs:
	$(DOCKER_COMPOSE) logs -f app

# ────────────────────────────────
# Build (native)
# ────────────────────────────────

## build: Compile binaries locally
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	$(GOBUILD) -o $(BUILD_DIR)/$(TOKEN_TOOL) ./pkg/token

## run: Run server locally (requires config/config.yaml)
run:
	$(GOCMD) run ./cmd/server -config config/config.yaml

# ────────────────────────────────
# Test & Lint
# ────────────────────────────────

## test: Run all tests with race detection
test:
	$(GOTEST) -v -race -cover ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	$(GOCMD) clean
	rm -rf $(BUILD_DIR)

# ────────────────────────────────
# Token generation
# ────────────────────────────────

## token: Generate a JWT token (requires built binary)
token:
	@test -f $(BUILD_DIR)/$(TOKEN_TOOL) || $(MAKE) build
	@echo "Usage:"
	@echo "  ./$(BUILD_DIR)/$(TOKEN_TOOL) --user <id> --email <email> --secret <jwt_secret> --expire 365d"
	@echo ""
	@echo "Example:"
	@echo "  ./$(BUILD_DIR)/$(TOKEN_TOOL) --user admin --email admin@example.com --expire 365d"

# ────────────────────────────────
# Help
# ────────────────────────────────

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ": "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' | \
		sed 's/^## //'
