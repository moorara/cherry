name := cherry
build_path := ./build
coverage_path := ./coverage

docker_tag ?= latest
docker_image ?= moorara/$(name)


clean:
	@ rm -rf *.log $(name) $(build_path) $(coverage_path)

run:
	@ go run main.go

build:
	@ ./scripts/build.sh --main main.go --binary $(name)

build-all:
	@ ./scripts/build.sh --all --main main.go --binary $(build_path)/$(name)

test:
	@ go test -race ./...

test-short:
	@ go test -short ./...

coverage:
	@ ./scripts/test-unit-cover.sh

docker:
	@ docker build -t $(docker_image):$(docker_tag) .

push:
	@ docker push $(docker_image):$(docker_tag)


.PHONY: clean
.PHONY: run build build-all
.PHONY: test test-short coverage
.PHONY: docker push
