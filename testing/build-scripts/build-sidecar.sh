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

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
MAKEDIR=$SCRIPTDIR/../../

make -C $MAKEDIR build.sidecar GOOS=linux GOARCH=amd64
STATUS=$?
if [ $STATUS -ne 0 ]; then
    echo -e "\n***********\nFAILED: make failed for sidecar.\n***********\n"
    exit $STATUS
fi

make -C $MAKEDIR dockerize.sidecar SIDECAR_IMAGE_NAME=amalgam8/a8-sidecar:alpine SIDECAR_DOCKERFILE=$MAKEDIR/docker/Dockerfile.sidecar.alpine
STATUS=$?
if [ $STATUS -ne 0 ]; then
    echo -e "\n***********\nFAILED: docker build failed for sidecar (alpine version)\n***********\n"
    exit $STATUS
fi
