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
EXAMPLESDIR=$SCRIPTDIR/../../examples

echo "Testing kubernetes-based deployment.."

sudo $EXAMPLESDIR/kubernetes/install-kubernetes.sh
sleep 5

$EXAMPLESDIR/kubernetes/run-controlplane-local-k8s.sh start
sleep 5

kubectl create -f $EXAMPLESDIR/kubernetes/gateway.yaml
sleep 5

kubectl create -f $SCRIPTDIR/../build-scripts/kubernetes.yaml
echo "Waiting for the services to come online.."
sleep 5

# Run the actual test workload
$SCRIPTDIR/demo_script.sh

echo "Kubernetes tests successful. Cleaning up.."
$EXAMPLESDIR/kubernetes/cleanup.sh
sleep 2
sudo $EXAMPLESDIR/kubernetes/uninstall-kubernetes.sh
