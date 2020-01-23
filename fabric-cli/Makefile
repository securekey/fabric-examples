#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Supported Targets:
#
# populate-sdk: retrieves all fabric-sdk-files required for tests
# fabric-sdk-go-pin: populates only upstream files (not included in git)
# example-network: start the example fabric network
# example-network-clean: stops any running network and cleans up the artifacts
# clean: removes all generated artifacts
#

# Tool commands (overridable)
GO_CMD             ?= go
GO_DEP_CMD         ?= dep
DOCKER_CMD         ?= docker
DOCKER_COMPOSE_CMD ?= docker-compose

# fabric-sdk-go version (overridable)
FABRIC_SDK_GO_BRANCH    ?= master
FABRIC_SDK_GO_COMMIT    ?= 8f3d32c9d1a66f88d405c9f5335870c5277f773a

# Test fixture paths
FIXTURE_DOCKERENV_PATH := fabric-sdk-go/test/fixtures/dockerenv

.PHONY: polulate-sdk
populate-sdk: fabric-sdk-go-pin fabric-sdk-go-populate

.PHONY: fabric-sdk-go-pin
fabric-sdk-go-pin:
	@echo "Pinning fabric-sdk-go ..."
	@UPSTREAM_COMMIT=$(FABRIC_SDK_GO_COMMIT) UPSTREAM_BRANCH=$(FABRIC_SDK_GO_BRANCH) scripts/fabric-sdk-go/apply_upstream.sh

.PHONY: fabric-sdk-go-populate
fabric-sdk-go-populate:
	@scripts/fabric-sdk-go/populate.sh

.PHONY: clean
clean: example-network-clean
	rm -Rf vendor
	rm -Rf fabric-sdk-go

.PHONY: example-network
example-network:
	@cd ./fabric-sdk-go && \
		make dockerenv-stable-up

.PHONY: example-network-clean
example-network-clean:
	@DOCKER_CMD=$(DOCKER_CMD) scripts/fabric-sdk-go/example-network-clean.sh
