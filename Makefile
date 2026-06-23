GO := $(shell pwd)/go/bin/go
GOPATH := $(shell pwd)/.gopath
export GOPATH

.PHONY: all server agent agent-windows clean run-server run-agent

all: server agent

## Build server binary
server:
	@echo "🔨 Building server..."
	@cd server && $(GO) build -o ../bin/workspace-server .
	@echo "✅ bin/workspace-server"

## Build agent binary (Linux/Ubuntu)
agent:
	@echo "🔨 Building agent (Linux)..."
	@cd agent && $(GO) build -o ../bin/workspace-agent .
	@echo "✅ bin/workspace-agent"

## Build agent binary for Windows (cross-compile)
agent-windows:
	@echo "🔨 Building agent (Windows)..."
	@cd agent && GOOS=windows GOARCH=amd64 $(GO) build -o ../bin/workspace-agent.exe .
	@echo "✅ bin/workspace-agent.exe"

## Run server (development)
run-server:
	@echo "🚀 Starting server..."
	@cd server && cp -n .env.example .env 2>/dev/null || true
	@cd server && $(GO) run .

## Run agent (development)
run-agent:
	@echo "🤖 Starting agent..."
	@cd agent && cp -n .env.example .env 2>/dev/null || true
	@cd agent && $(GO) run .

## Clean binaries
clean:
	@rm -rf bin/
	@echo "🗑️  Cleaned"

## Tidy all modules
tidy:
	@cd server && $(GO) mod tidy
	@cd agent && $(GO) mod tidy
	@echo "✅ Modules tidied"

## Show help
help:
	@echo "Workspace Commander — Makefile"
	@echo ""
	@echo "  make server          Build server binary"
	@echo "  make agent           Build agent binary (Linux)"
	@echo "  make agent-windows   Build agent binary (Windows)"
	@echo "  make run-server      Run server (dev)"
	@echo "  make run-agent       Run agent (dev)"
	@echo "  make all             Build both"
	@echo "  make clean           Remove binaries"
