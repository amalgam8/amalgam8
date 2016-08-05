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
set -o errexit
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

#git clone https://github.com/amalgam8/controller ../controller
#git clone https://github.com/amalgam8/registry ../registry
#git clone https://github.com/amalgam8/sidecar ../sidecar
git clone https://github.com/amalgam8/examples
#sudo pip install git+https://github.com/amalgam8/a8ctl

$SCRIPTDIR/build-scripts/build-amalgam8.sh
$SCRIPTDIR/examples/docker/run-controlplane-docker.sh start
sleep 5
docker-compose -f $SCRIPTDIR/examples/docker/gateway.yaml up -d
sleep 5
docker-compose -f $SCRIPTDIR/examples/docker/bookinfo.yaml up -d
sleep 10
$SCRIPTDIR/testing/demo_script.sh
