.PHONY: build run dev test test-pkg lint fmt tidy sqlc tailwind tailwind-watch docker-build docker-run docker-up docker-down help

BINARY=rustyfinancial

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build            compile binary"
	@echo "  run              go run ./cmd/server"
	@echo "  dev              hot reload (air + bun css:watch)"
	@echo "  test             run all tests"
	@echo "  test-pkg PKG=... run tests for a single package"
	@echo "  lint             golangci-lint"
	@echo "  fmt              go fmt ./..."
	@echo "  tidy             go mod tidy"
	@echo "  sqlc             regenerate internal/db/ from sql/"
	@echo "  tailwind         build CSS once"
	@echo "  tailwind-watch   rebuild CSS on change"
	@echo "  docker-build     build Docker image"
	@echo "  docker-run       run Docker image (foreground)"
	@echo "  docker-up        run Docker image (detached)"
	@echo "  docker-down      stop and remove containers"

build:
	go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

dev: node_modules
	trap 'kill 0' SIGINT; \
	bun run css:watch & \
	go tool air; \
	wait

test:
	go test ./...

test-pkg:
	go test ./$(PKG)/...

lint:
	go tool golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy

sqlc:
	go tool sqlc generate

node_modules:
	bun install

tailwind: node_modules
	bun run css

tailwind-watch: node_modules
	bun run css:watch

docker-build:
	docker compose build

docker-run:
	docker compose up

docker-up:
	docker compose up -d

docker-down:
	docker compose down
