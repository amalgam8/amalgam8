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
export A8_CONTAINER_ENV="k8s"

if [ "$A8_TEST_SUITE" == "examples" ]; then
    export A8_TEST_ENV="examples"
else
    export A8_TEST_ENV="testing"
fi

# Increase memory limit for elasticsearch 5.1
sudo sysctl -w vm.max_map_count=262144

echo "Testing kubernetes-based deployment.."

sudo $SCRIPTDIR/install-kubernetes.sh
sleep 10

echo "Starting control plane"
kubectl create -f $SCRIPTDIR/controlplane.yaml
sleep 10

sed -e "s/{A8_TEST_ENV}/$A8_TEST_ENV/" $SCRIPTDIR/bookinfo.yaml | kubectl create -f -
echo "Waiting for the services to come online.."
sleep 10

# Run the actual test workload
$SCRIPTDIR/../test-scripts/bookinfo.sh $A8_TEST_SUITE

echo "Kubernetes tests successful."
echo "Cleaning up Bookinfo apps.."
sed -e "s/{A8_TEST_ENV}/$A8_TEST_ENV/" $SCRIPTDIR/bookinfo.yaml | kubectl delete -f - || echo "Probably already down"
sleep 5

if [ "$A8_TEST_SUITE" == "examples" ]; then
	kubectl create -f $SCRIPTDIR/helloworld.yaml
	sleep 10

	$SCRIPTDIR/../test-scripts/helloworld.sh $A8_TEST_SUITE

	kubectl delete -f $SCRIPTDIR/helloworld.yaml
	sleep 5
fi

echo "Stopping control plane services..."
kubectl delete -f $SCRIPTDIR/controlplane.yaml

sleep 5
sudo $SCRIPTDIR/uninstall-kubernetes.sh
