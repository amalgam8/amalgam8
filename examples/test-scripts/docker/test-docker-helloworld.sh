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

echo "Testing docker-based deployment.."

export A8_REGISTRY_URL=http://localhost:31300
export A8_CONTROLLER_URL=http://localhost:31200
export GATEWAY_URL=http://localhost:32000

$SCRIPTDIR/run-controlplane-docker.sh start
sleep 5

docker-compose -f $SCRIPTDIR/docker-helloworld.yaml up -d
sleep 2
docker-compose -f $SCRIPTDIR/docker-helloworld.yaml scale helloworld-v1=2
sleep 2
docker-compose -f $SCRIPTDIR/docker-helloworld.yaml scale helloworld-v2=2
sleep 10

# Make sure if the testcase fails, the cleanup still runs
set +e

# Run the actual test workload
$SCRIPTDIR/../helloworld.sh "docker"
RC=$?

#echo "Docker tests completed with return code: $RC. Cleaning up.."
#$SCRIPTDIR/cleanup-helloworld.sh $APP_VER

#sleep 10

exit $RC
