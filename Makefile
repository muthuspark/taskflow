.PHONY: help build build-frontend build-backend run dev clean test

help:
	@echo "TaskFlow - Task Scheduler & Runner"
	@echo ""
	@echo "Available commands:"
	@echo "  make build           Build frontend and backend"
	@echo "  make build-frontend  Build Vue.js frontend"
	@echo "  make build-backend   Build Go backend"
	@echo "  make run             Run the backend server"
	@echo "  make dev             Run backend in development mode"
	@echo "  make dev-frontend    Run frontend dev server"
	@echo "  make test            Run tests"
	@echo "  make clean           Clean build artifacts"

build: build-frontend build-backend
	@echo "Build complete"

build-frontend:
	@echo "Building frontend..."
	cd web/frontend && npm install && npm run build
	@echo "Frontend build complete"

build-backend:
	@echo "Building backend..."
	go build -o bin/taskflow ./cmd/taskflow
	@echo "Backend build complete"

run: build
	@echo "Starting TaskFlow..."
	export JWT_SECRET=$$(openssl rand -hex 32) && ./bin/taskflow

dev:
	@echo "Starting TaskFlow in development mode..."
	export JWT_SECRET=$$(openssl rand -hex 32) && PORT=8080 go run ./cmd/taskflow

dev-frontend:
	@echo "Starting frontend dev server..."
	cd web/frontend && npm install && npm run dev

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf web/frontend/dist/
	rm -rf web/frontend/node_modules/
