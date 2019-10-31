name := cherry
docker_image ?= moorara/$(name)
docker_tag ?= latest


build:
	@ cherry build -cross-compile=false

build-all:
	@ cherry build -cross-compile=true

test:
	@ go test -race ./...

test-short:
	@ go test -short ./...

coverage:
	@ go test -covermode=atomic -coverprofile=c.out ./...
	@ go tool cover -html=c.out -o coverage.html

docker:
	@ docker image build -t $(docker_image):$(docker_tag) .

push:
	@ docker image push $(docker_image):$(docker_tag)

push-latest:
	@ docker image tag $(docker_image):$(docker_tag) $(docker_image):latest
	  docker image push $(docker_image):latest

save-docker:
	@ docker image save -o docker.tar $(docker_image):$(docker_tag)

load-docker:
	@ docker image load -i docker.tar


.PHONY: build build-all
.PHONY: test test-short coverage
.PHONY: docker push push-latest save-docker load-docker
