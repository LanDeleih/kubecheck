NAME := kubecheck
DIR := bin
GIT_SHA = $(shell git rev-list --count HEAD)-$(shell git rev-parse --short=7 HEAD)
COMMIT_TIME = $(shell git show --format=%ct --no-patch)
VERSION = $(shell git describe --tags `git rev-list --tags --max-count=1`)
LFLAGS ?= -X main.gitsha=${GIT_SHA} -X main.committed=${COMMIT_TIME} -X main.VERSION=${VERSION}

.PHONY: clean

default: build

install_deps:
	go mod vendor

clean:
	rm -rf bin/

build: install_deps
	@echo "--> Compiling the project"
	go build -ldflags "${LFLAGS}" -o $(DIR)/$(NAME) ./app/main.go