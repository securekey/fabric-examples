#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code used in the examples originating from the fabric-sdk-go project

set -e

UPSTREAM_PROJECT="github.com/hyperledger/fabric-sdk-go"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH}"
SCRIPTS_PATH="scripts/fabric-sdk-go"

FABRIC_SDK_GO_PATH='fabric-sdk-go'

####
# Clone and patch packages into repo

# Clone fabric-sdk-go project into temporary directory
echo "Fetching upstream project ($UPSTREAM_PROJECT:$UPSTREAM_COMMIT) ..."
CWD=`pwd`
TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`

TMP_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
mkdir -p $TMP_PROJECT_PATH
cd ${TMP_PROJECT_PATH}/..

git clone https://${UPSTREAM_PROJECT}.git
cd $TMP_PROJECT_PATH
git checkout $UPSTREAM_BRANCH
git reset --hard $UPSTREAM_COMMIT

cd $CWD

echo 'Removing current upstream project from working directory ...'
rm -Rf "${FABRIC_SDK_GO_PATH}"
mkdir -p "${FABRIC_SDK_GO_PATH}"

declare -a PATHS=(
    "test/fixtures/*"
)

for i in "${PATHS[@]}"
do
    DIR=$(dirname "${i}")
    DEST=$FABRIC_SDK_GO_PATH/$DIR
    mkdir -p $DEST
    cp -R $TMP_PROJECT_PATH/${i} $DEST
done

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
