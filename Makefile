# Image URL to use all building/pushing image targets
IMG ?= alexdecb/l2sm-switch:2.7.1

# ENV variables for sample-config
CONTROLLERIP ?= $(shell grep CONTROLLERIP .env | cut -d '=' -f2)
NODENAME ?= $(shell grep NODENAME .env | cut -d '=' -f2)
NEDNAME ?= brtun

# CONTAINER_TOOL defines the container tool to be used for building images.
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} -f ./build/Dockerfile .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

##@ Development

.PHONY: generate-proto
generate-proto: ## Generate gRPC code from .proto file.
	protoc -I=api/v1 --go_out=paths=source_relative:./pkg/nedpb --go-grpc_out=paths=source_relative:./pkg/nedpb api/v1/ned.proto

.PHONY: sample-config
sample-config: ## Create sample config.json if it doesn't exist.
	echo "{\"ConfigDir\":\"\",\"ControllerIP\":\"$(CONTROLLERIP)\",\"NodeName\":\"$(NODENAME)\",\"NedName\":\"$(NEDNAME)\"}" > config/config.json; \
	echo "{\"name\":\"$(NODENAME)\",\"nodeIP\":\"$(CONTROLLERIP)\",\"neighborNodes\":[\"\"]}" > config/neighbors.json; \
##@ Run

.PHONY: run
run: ## Run the application.
	sudo /usr/local/go/bin/go run ./cmd/ned-server --config_dir ./config/config.json --neighbors_dir ./config/neighbors.json

##@ OVS

.PHONY: deploy-ovs
deploy-ovs: ## Deploy OVS server and switch.
	ovsdb-server --remote=punix:/var/run/openvswitch/db.sock --remote=db:Open_vSwitch,Open_vSwitch,manager_options --pidfile=/var/run/openvswitch/ovsdb-server.pid --detach
	ovs-vsctl --db=unix:/var/run/openvswitch/db.sock --no-wait init
	ovs-vswitchd --pidfile=/var/run/openvswitch/ovs-vswitchd.pid --detach

.PHONY: clean
clean: ## Clean up OVS bridge.
	sudo ovs-vsctl del-br brtun
