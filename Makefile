REGISTRY=quay.io
REPOSITORY=metaconflux
IMAGE=api
VERSION=v0.0.1

TAG=$(REGISTRY)/$(REPOSITORY)/$(IMAGE):$(VERSION)

server:
	go run cmd/server/main.go

build:
	CGO_ENABLED=0 go build -o _build/server cmd/server/main.go

build-container:
	podman build -t $(TAG) .