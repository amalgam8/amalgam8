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

A8_TEST_SUITE=$1

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Set env vars
export A8_CONTROLLER_URL=http://localhost:31200
export A8_REGISTRY_URL=http://localhost:31300
export A8_GATEWAY_URL=http://localhost:32000
export A8_LOG_SERVER=http://localhost:30200
export A8_GREMLIN_URL=http://localhost:31500

# The test script checks this var to determine if we're using docker or k8s
export A8_CONTAINER_ENV="docker"

if [ "$A8_TEST_SUITE" == "examples" ]; then
    export A8_TEST_ENV="examples"
else
    export A8_TEST_ENV="testing"
fi

# Make sure there are no containers running
docker-compose -f $SCRIPTDIR/helloworld.yaml kill
docker-compose -f $SCRIPTDIR/helloworld.yaml rm -f
docker-compose -f $SCRIPTDIR/bookinfo.yaml kill
docker-compose -f $SCRIPTDIR/bookinfo.yaml rm -f
docker-compose -f $SCRIPTDIR/controlplane.yaml kill
docker-compose -f $SCRIPTDIR/controlplane.yaml rm -f
sleep 5

# Increase memory limit for elasticsearch 5.1
sudo sysctl -w vm.max_map_count=262144

echo "Testing docker-based deployment.."

echo "starting Control plane components (registry, and controller)"
docker-compose -f $SCRIPTDIR/controlplane.yaml up -d
echo "waiting for the cluster to initialize.."

sleep 5

if [ "$A8_TEST_SUITE" == "examples" ]; then
	docker-compose -f $SCRIPTDIR/helloworld.yaml up -d
	sleep 10

	# Run the actual test workload
	$SCRIPTDIR/../test-scripts/helloworld.sh $A8_TEST_SUITE

	docker-compose -f $SCRIPTDIR/helloworld.yaml kill
	docker-compose -f $SCRIPTDIR/helloworld.yaml rm -f
	sleep 10
fi

docker-compose -f $SCRIPTDIR/bookinfo.yaml up -d
sleep 10

# Run the actual test workload
$SCRIPTDIR/../test-scripts/bookinfo.sh $A8_TEST_SUITE

echo "Docker tests successful."
echo "Cleaning up Bookinfo apps.."
docker-compose -f $SCRIPTDIR/bookinfo.yaml kill
docker-compose -f $SCRIPTDIR/bookinfo.yaml rm -f

echo "Stopping control plane services..."
docker-compose -f $SCRIPTDIR/controlplane.yaml kill
docker-compose -f $SCRIPTDIR/controlplane.yaml rm -f

sleep 10
