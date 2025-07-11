.PHONY: build

# Development

build_rev := "main"
ifneq ($(wildcard .git),)
	build_rev := $(shell git rev-parse --short HEAD)
endif
build_date := $(shell date -u '+%Y-%m-%dT%H:%M:%S')

setup:
	@go mod download

lint:
	@golangci-lint run ./...
	@echo "✓ lint"

vet:
	@go vet ./...
	@echo "✓ vet"

test:
	@go test ./...
	@echo "✓ test"

build:
	@go build -ldflags "-X main.commit=$(build_rev) -X main.date=$(build_date)" -o build/codapi -v cmd/main.go

run:
	@./build/codapi


# Containers

image:
	@[ -n "$(name)" ] || (echo "Syntax: make image name=<image-name>" >&2; exit 1)
	@echo "Building image codapi/$(name)"
	@docker build --file sandboxes/$(name)/Dockerfile --tag codapi/$(name):latest sandboxes/$(name)/
	@echo "✓ codapi/$(name)"

network:
	docker network create --internal codapi

# Host OS

mount-tmp:
	mount -t tmpfs tmpfs /tmp -o rw,exec,nosuid,nodev,size=64m,mode=1777

# Deployment

app-download:
	@curl -L -o codapi.zip "https://api.github.com/repos/nalgeon/codapi/actions/artifacts/$(id)/zip"
	@unzip -ou codapi.zip
	@chmod +x build/codapi
	@rm -f codapi.zip
	@echo "OK"

app-start:
	@nohup build/codapi > codapi.log 2>&1 & echo $$! > codapi.pid
	@echo "started codapi"

app-stop:
	@kill $(shell cat codapi.pid)
	@rm -f codapi.pid
	@echo "stopped codapi"
