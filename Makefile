name := cherry
docker_image ?= moorara/$(name)
docker_tag ?= latest


clean:
	@ rm -rf bin coverage $(name)

run:
	@ go run main.go

build:
	@ cherry build -cross-compile=false

build-all:
	@ cherry build -cross-compile=true

test:
	@ go test -race ./...

test-short:
	@ go test -short ./...

coverage:
	@ cherry test

docker:
	@ docker build -t $(docker_image):$(docker_tag) .

push:
	@ docker push $(docker_image):$(docker_tag)

save-docker:
	@ docker image save -o docker.tar $(docker_image):$(docker_tag)

load-docker:
	@ docker image load -i docker.tar


.PHONY: clean
.PHONY: run build build-all
.PHONY: test test-short coverage
.PHONY: docker push save-docker load-docker
