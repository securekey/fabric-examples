#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script deletes all docker containers, and example network docker images

set -e

cd fabric-sdk-go/test/fixtures/dockerenv/
docker-compose -f docker-compose.yaml -f docker-compose-chaincoded.yaml down
cd ../../../..
