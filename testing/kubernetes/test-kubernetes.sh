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

# Increase memory limit for elasticsearch 5.1
sudo sysctl -w vm.max_map_count=262144

echo "Testing kubernetes-based deployment.."

sudo $SCRIPTDIR/install-kubernetes.sh
sleep 10

echo "Starting control plane"
kubectl create -f $SCRIPTDIR/controlplane.yaml
sleep 10

kubectl create -f $SCRIPTDIR/bookinfo.yaml
echo "Waiting for the services to come online.."
sleep 10

# Run the actual test workload
$SCRIPTDIR/../test-scripts/demo_script.sh

echo "Kubernetes tests successful."
echo "Cleaning up Bookinfo apps.."
kubectl delete -f $SCRIPTDIR/bookinfo.yaml
sleep 5

echo "Stopping control plane services..."
kubectl delete -f $SCRIPTDIR/controlplane.yaml

sleep 5
sudo $SCRIPTDIR/uninstall-kubernetes.sh
