REGISTRY=quay.io
REPOSITORY=metaconflux
IMAGE=api
VERSION=v0.0.2

CONTAINER_NAME=api

TAG=$(REGISTRY)/$(REPOSITORY)/$(IMAGE):$(VERSION)

server:
	go run cmd/server/main.go

build:
	go build -o _build/server cmd/server/main.go

build-container:
	podman build -t $(TAG) .

push-container:
	podman push $(TAG)

login:
	docker login -u $(QUAY_USER) -p $(QUAY_PASSWORD) $(REGISTRY)

release-container: login build-container push-container

pull-image:
	podman pull $(TAG)

deploy: pull-image
	podman stop $(CONTAINER_NAME);\
	podman rm $(CONTAINER_NAME);\
	podman run -d --name $(CONTAINER_NAME)\
			 -p 8081:8081\
			 -v ${PWD}/config.yaml:/opt/metaconflux/config.yaml:z\
			 -v ${PWD}/gorm.db:/opt/metaconflux/gorm.db:z\
			 $(TAG)
