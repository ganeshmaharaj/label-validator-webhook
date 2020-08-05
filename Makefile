.DEFAULT_GOAL: build
VER:=$(shell git describe --tags HEAD)

build: export DOCKER_BUILDKIT = 1
build:
	docker build -t gmmaha/labeling-validator:$(VER) .
	docker tag gmmaha/labeling-validator:$(VER) gmmaha/labeling-validator:latest

push:
	$(foreach var,$(VER) latest, \
		docker push gmmaha/labeling-validator:$(var);)
