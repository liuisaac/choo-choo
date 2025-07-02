# === VARIABLES ===
APP_NAME=choochoo
CLI_NAME=choochoo-shell
SERVER_DIR=./cmd/server
SHELL_DIR=./cmd/shell
BUILD_DIR=./bin

# === DEFAULT ===
all: build

# === BUILD TARGETS ===
build: build-server build-cli

build-server:
	@echo "🔨 Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(SERVER_DIR)

build-cli:
	@echo "🔨 Building $(CLI_NAME)..."
	@go build -o $(BUILD_DIR)/$(CLI_NAME) $(SHELL_DIR)

# === RUN TARGETS ===
run: build-server
	@echo "🚀 Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

shell: build-cli
	@echo "💻 Launching $(CLI_NAME)..."
	@$(BUILD_DIR)/$(CLI_NAME)

# === CLEAN ===
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)

# === TEST ===
test:
	@echo "🧪 Running tests..."
	@go test ./...

.PHONY: all build build-server build-cli run shell clean test
