.DEFAULT_GOAL: build
VER:=$(shell git describe --tags HEAD)

build: export DOCKER_BUILDKIT = 1
build:
	docker build -t gmmaha/labeling-validator:$(VER) .

push:
	docker tag gmmaha/labeling-validator:$(VER) gmmaha/labeling-validator:latest
	$(foreach var,$(VER) latest, \
		docker push gmmaha/labeling-validator:$(var);)
