GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
IMAGE ?= inspector:$(GIT_COMMIT)
KIND_CLUSTER ?= kind-muyiwa-dev
MINIKUBE_PROFILE ?= minikube

# detect host arch for macOS builds
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	GOARCH_MAC := amd64
else ifeq ($(UNAME_M),arm64)
	GOARCH_MAC := arm64
else
	GOARCH_MAC := amd64
endif

.PHONY: help build build-linux build-macos build-all docker-build create-kind kind-load minikube-build deploy rollout port-forward run fmt tidy test

help:
	@echo "Targets:"
	@echo "  make build           Build linux binary with embedded metadata"
	@echo "  make build-macos     Build macOS binary(s) (arch-aware)"
	@echo "  make build-all       Build linux and macOS binaries"
	@echo "  make docker-build    Build docker image with build metadata (uses current docker)"
	@echo "  make create-kind     Create a local kind cluster named $(KIND_CLUSTER)"
	@echo "  make kind-load       Load built image into kind cluster ($(KIND_CLUSTER))"
	@echo "  make minikube-build  Build image into minikube's docker daemon"
	@echo "  make deploy          Apply k8s manifests and set image to $(IMAGE)"
	@echo "  make rollout         Wait for rollout to finish"
	@echo "  make port-forward    Port-forward inspector to localhost:3000"
	@echo "  make run             Run service locally (uses KUBECONFIG)"
	@echo "  make fmt             Run gofmt"
	@echo "  make tidy            Run go mod tidy"
	@echo "  make test            Run go test ./..."

build: build-linux build-macos-arm

# Build Linux binary (container runtime target) -> build/linux/inspector
build-linux:
	@mkdir -p build/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags "-s -w -X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'" \
		-o build/linux/inspector ./cmd/inspector

# Build macOS binary for Apple Silicon (M1) -> bin/arm/inspector
build-macos-arm:
	@mkdir -p bin/arm
	GOOS=darwin GOARCH=arm64 \
		go build -ldflags "-s -w -X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'" \
		-o bin/arm/inspector ./cmd/inspector

# Build both linux and macOS-arm
build-all: build-linux build-macos-arm

docker-build:
	docker build --build-arg BUILD_TIME="$(BUILD_TIME)" --build-arg GIT_COMMIT="$(GIT_COMMIT)" -t $(IMAGE) .


minikube-build:
	@echo "Building image inside minikube docker daemon ($(MINIKUBE_PROFILE))"
	@eval "$$($(MINIKUBE_PROFILE) docker-env)" && docker build --build-arg BUILD_TIME="$(BUILD_TIME)" --build-arg GIT_COMMIT="$(GIT_COMMIT)" -t $(IMAGE) .

# ===========================

create-kind:
	kind create cluster --name $(KIND_CLUSTER)

kind-load: docker-build
	kind load docker-image $(IMAGE) --name $(KIND_CLUSTER)

deploy:
	kubectl apply -f k8s/rbac.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl set image deployment/inspector inspector=$(IMAGE) --record

rollout:
	kubectl rollout status deployment/inspector

port-forward:
	kubectl port-forward deploy/inspector 3000:3000

# ===========================


run:
	export KUBECONFIG=$$HOME/.kube/config && go run ./cmd/inspector

fmt:
	gofmt -w .

tidy:
	go mod tidy

test:
	go test ./...
