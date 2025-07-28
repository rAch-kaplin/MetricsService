PROJECT_NAME := MetricsService

SERVER_SRC_DIR := ./cmd/server
AGENT_SRC_DIR := ./cmd/agent

SERVER_BIN_NAME := server
AGENT_BIN_NAME := agent

SERVER_FULL_PATH := $(SERVER_SRC_DIR)/$(SERVER_BIN_NAME)
AGENT_FULL_PATH := $(AGENT_SRC_DIR)/$(AGENT_BIN_NAME)

.PHONY: all server agent build clean test

all: build

build: server agent

server: deps
	@go build -o $(SERVER_FULL_PATH) $(SERVER_SRC_DIR)
	@echo "Built $(SERVER_FULL_PATH)"

agent: deps
	@go build -o $(AGENT_FULL_PATH) $(AGENT_SRC_DIR)
	@echo "Built $(AGENT_FULL_PATH)"

test:
	@go test ./... -v

bench:
	@go test -bench=. -benchmem ./...

deps:
	@go mod download

lint: deps
	@golangci-lint run

clean:
	@rm -f $(SERVER_FULL_PATH) $(AGENT_FULL_PATH)
	@echo "Cleaned."

