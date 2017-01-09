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
set -e

APP_VER=$1
if [ -z $APP_VER ]; then
        APP_VER="latest"
fi
export APP_VER=$APP_VER

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

export A8_REGISTRY_URL=http://localhost:31300
export A8_CONTROLLER_URL=http://localhost:31200
export GATEWAY_URL=http://localhost:32000

docker-compose -f $SCRIPTDIR/test-scripts/docker/docker-bookinfo.yaml up -d
sleep 2
