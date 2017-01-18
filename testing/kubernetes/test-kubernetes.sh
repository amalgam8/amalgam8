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
export A8_CONTAINER_ENV="k8s"

startup_pods(){
    if [ "$A8_TEST_SUITE" == "examples" ]; then
        kubectl create -f $1
    else
        sed '
        s/${A8_TEST_ENV}/testing/g
        s/${A8_RELEASE}/latest/g
	' $1 | kubectl create -f -
    fi
}

shutdown_pods(){
    if [ "$A8_TEST_SUITE" == "examples" ]; then
        kubectl delete -f $1 || echo "Probably already down"
    else
        sed '
        s/${A8_TEST_ENV}/testing/g
        s/${A8_RELEASE}/latest/g
	' $1 | kubectl delete  -f - || echo "Probably already down"
    fi
}

if [ "$A8_TEST_SUITE" == "examples" ]; then
    if [ -z "$A8_RELEASE" ]; then
        echo "Release must be specified if running examples test suite."
        exit 1
    fi
    echo "======= Running the examples test suite ======="
    # Generate the examples yaml file
    $SCRIPTDIR/../generate_example_yaml.sh $A8_RELEASE
    export A8_TEST_ENV="examples"
    HELLOWORLD_YAML=$SCRIPTDIR/../../examples/k8s-helloworld.yaml
    BOOKINFO_YAML=$SCRIPTDIR/../../examples/k8s-bookinfo.yaml
    CONTROLPLANE_YAML=$SCRIPTDIR/../../examples/k8s-controlplane.yaml
else
    echo "======= Running the integration test suite ======="
    export A8_TEST_ENV="testing"
    HELLOWORLD_YAML=$SCRIPTDIR/helloworld.yaml
    BOOKINFO_YAML=$SCRIPTDIR/bookinfo.yaml
    CONTROLPLANE_YAML=$SCRIPTDIR/controlplane.yaml
fi

# Make sure kubernetes is not active
shutdown_pods $HELLOWORLD_YAML
shutdown_pods $BOOKINFO_YAML
shutdown_pods $CONTROLPLANE_YAML
sleep 5

# Increase memory limit for elasticsearch 5.1
sudo sysctl -w vm.max_map_count=262144

echo "Testing kubernetes-based deployment.."

sudo $SCRIPTDIR/install-kubernetes.sh
sleep 10

echo "Starting control plane"
startup_pods $CONTROLPLANE_YAML
sleep 10

startup_pods $BOOKINFO_YAML
echo "Waiting for the services to come online.."
sleep 10

# Run the actual test workload
$SCRIPTDIR/../test-scripts/bookinfo.sh $A8_TEST_SUITE

echo "Kubernetes tests successful."
echo "Cleaning up Bookinfo apps.."
shutdown_pods $BOOKINFO_YAML || echo "Probably already down"
sleep 5

if [ "$A8_TEST_SUITE" == "examples" ]; then
	startup_pods $HELLOWORLD_YAML
	sleep 10

	$SCRIPTDIR/../test-scripts/helloworld.sh $A8_TEST_SUITE

	shutdown_pods $HELLOWORLD_YAML
	sleep 5
fi

echo "Stopping control plane services..."
shutdown_pods $CONTROLPLANE_YAML

sleep 5
sudo $SCRIPTDIR/uninstall-kubernetes.sh
