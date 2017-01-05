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

APP_VER=$1
if [ -z $APP_VER ]; then
        APP_VER="latest"
fi

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "Testing kubernetes-based deployment.."

sudo $SCRIPTDIR/install-kubernetes.sh
sleep 10

export A8_REGISTRY_URL=http://localhost:31300
export A8_CONTROLLER_URL=http://localhost:31200
export GATEWAY_URL=http://localhost:32000

$SCRIPTDIR/run-controlplane-kubernetes.sh start
sleep 10

sed -e "s/{APP_VER}/$APP_VER/" $SCRIPTDIR/k8s-bookinfo.yaml | kubectl create -f -
echo "Waiting for the services to come online.."
sleep 60

# Make sure the cleanup is run if the testcase fails
set +e

# Run the actual test workload
$SCRIPTDIR/../bookinfo.sh "k8s"
RC=$?

echo "Kubernetes tests completed with return code: $RC. Cleaning up.."
$SCRIPTDIR/cleanup-bookinfo.sh $APP_VER
sleep 5
sudo $SCRIPTDIR/uninstall-kubernetes.sh

exit $RC
