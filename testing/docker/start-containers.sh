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
A8_RELEASE=$2

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
    if [ -z "$A8_RELEASE" ]; then
        echo "Release must be specified if running examples test suite."
	exit 1
    fi
    echo "======= Running the examples test suite ======="
    # Generate the examples yaml file
    $SCRIPTDIR/../generate_example_yaml.sh $A8_RELEASE
    export A8_TEST_ENV="examples"
    HELLOWORLD_YAML=$SCRIPTDIR/../../examples/docker-helloworld.yaml
    BOOKINFO_YAML=$SCRIPTDIR/../../examples/docker-bookinfo.yaml
    CONTROLPLANE_YAML=$SCRIPTDIR/../../examples/docker-controlplane.yaml
else
    echo "======= Running the integration test suite ======="
    export A8_TEST_ENV="testing"
    export A8_RELEASE="latest"
    HELLOWORLD_YAML=$SCRIPTDIR/helloworld.yaml
    BOOKINFO_YAML=$SCRIPTDIR/bookinfo.yaml
    CONTROLPLANE_YAML=$SCRIPTDIR/controlplane.yaml
fi

# Make sure there are no containers running
docker-compose -f $HELLOWORLD_YAML kill
docker-compose -f $HELLOWORLD_YAML rm -f
docker-compose -f $BOOKINFO_YAML kill
docker-compose -f $BOOKINFO_YAML rm -f
docker-compose -f $CONTROLPLANE_YAML kill
docker-compose -f $CONTROLPLANE_YAML rm -f
sleep 5

# Increase memory limit for elasticsearch 5.1
sudo sysctl -w vm.max_map_count=262144

echo "Testing docker-based deployment.."

echo "starting Control plane components (registry, and controller)"
docker-compose -f $CONTROLPLANE_YAML up -d
echo "waiting for the cluster to initialize.."

sleep 5

if [ "$A8_TEST_SUITE" == "examples" ]; then
	docker-compose -f $HELLOWORLD_YAML up -d
	sleep 10

	# Run the actual test workload
	$SCRIPTDIR/../test-scripts/helloworld.sh $A8_TEST_SUITE

	docker-compose -f $HELLOWORLD_YAML kill
	docker-compose -f $HELLOWORLD_YAML rm -f
	sleep 10
fi

docker-compose -f $BOOKINFO_YAML up -d
sleep 10
