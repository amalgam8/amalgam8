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
EXAMPLESDIR=$GOPATH/src/github.com/amalgam8/examples
EXAMPLESREPO=https://github.com/amalgam8/examples

#from https://gist.github.com/nicferrier/2277987
LOCALREPO=$EXAMPLESDIR

# We do it this way so that we can abstract if from just git later on
LOCALREPO_VC_DIR=$EXAMPLESREPO/.git

if [ ! -d $LOCALREPO_VC_DIR ]
then
    git clone $EXAMPLESREPO $EXAMPLESDIR
else
    cd $EXAMPLESDIR && git pull $EXAMPLESREPO
fi

# End

$SCRIPTDIR/build-scripts/build-amalgam8.sh

#######Test Docker setup
echo "Testing docker-based deployment.."
$EXAMPLESDIR/docker/run-controlplane-docker.sh start
sleep 5
docker-compose -f $EXAMPLESDIR/docker/gateway.yaml up -d
sleep 5
docker-compose -f $EXAMPLESDIR/docker/bookinfo.yaml up -d
sleep 10
$SCRIPTDIR/testing/demo_script.sh
echo "Docker tests successful. Cleaning up.."
$EXAMPLESDIR/docker/cleanup.sh
sleep 10


#######Test Kubernetes setup
echo "Testing kubernetes-based deployment.."
sudo $EXAMPLESDIR/kubernetes/install-kubernetes.sh
sleep 15
$EXAMPLESDIR/kubernetes/run-controlplane-local-k8s.sh start
sleep 15
kubectl create -f $EXAMPLESDIR/kubernetes/gateway.yaml
sleep 15
kubectl create -f $EXAMPLESDIR/kubernetes/bookinfo.yaml
echo "Waiting for the services to come online.."
sleep 60
$SCRIPTDIR/testing/demo_script.sh
echo "Kubernetes tests successful. Cleaning up.."
$EXAMPLESDIR/kubernetes/cleanup.sh
sleep 5
sudo $EXAMPLESDIR/kubernetes/uninstall-kubernetes.sh
