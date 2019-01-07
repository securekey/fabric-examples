#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script populates the vendor folder.

set -e

cd fabric-sdk-go

echo "Populating dockerd vendor ..."
declare chaincodedPath="scripts/_go/src/chaincoded"
rm -Rf ${chaincodedPath}/vendor/
mkdir -p ${chaincodedPath}/vendor/github.com/hyperledger/fabric
git clone --branch release-1.3 --depth=1 https://github.com/hyperledger/fabric.git ${chaincodedPath}/vendor/github.com/hyperledger/fabric

cd ..
