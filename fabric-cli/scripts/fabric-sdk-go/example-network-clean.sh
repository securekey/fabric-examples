#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script deletes all docker containers, and example network docker images

set -e

# TODO - delete example network containers
#CONTAINERS=$($DOCKER_CMD ps -a -q)
#if [[ ! -z $CONTAINERS ]]
#then
#	$DOCKER_CMD stop $CONTAINERS
#	$DOCKER_CMD rm $CONTAINERS
#fi

# TODO - delete examplecc images only
#EXAMPLE_IMAGES=$($DOCKER_CMD images | grep examplecc)
#if [[ ! -z $EXAMPLE_IMAGES ]]
#then
#	$DOCKER_CMD rmi $EXAMPLE_IMAGES
#fi

# For now, print how to do it manually
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cat $DIR/readme-example-network-clean.txt