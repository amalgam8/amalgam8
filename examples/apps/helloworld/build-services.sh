#!/bin/bash
#
# Copyright 2016 IBM Corporation
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


set -x

APP_VER=$1
if [ -z $APP_VER ]; then
  echo "The version must be set."
  exit 1
fi

if [ "$APP_VER" == "unknown" ]; then
  echo "The version cannot be unknown."
  exit 1
fi

A8_SIDECAR_RELEASE=$2

# Remove the v from the version
APP_VER=$(echo $APP_VER | sed "s/v//")

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

A8_SIDECAR_TAR_NAME=${A8_SIDECAR_RELEASE}.tar.gz
A8_SIDECAR_TAR=$SCRIPTDIR/../../../release/$A8_SIDECAR_TAR_NAME

#################################################################################
# Build the helloworld image
#################################################################################
docker build -t amalgam8/a8-examples-helloworld-v1:$APP_VER $SCRIPTDIR
docker build -t amalgam8/a8-examples-helloworld-v2:$APP_VER $SCRIPTDIR

cp $A8_SIDECAR_TAR $SCRIPTDIR
docker build -t amalgam8/a8-examples-helloworld-sidecar-v1:$APP_VER --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f $SCRIPTDIR/Dockerfile.sidecar $SCRIPTDIR
docker build -t amalgam8/a8-examples-helloworld-sidecar-v2:$APP_VER --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f $SCRIPTDIR/Dockerfile.sidecar $SCRIPTDIR
rm $SCRIPTDIR/$A8_SIDECAR_TAR_NAME
