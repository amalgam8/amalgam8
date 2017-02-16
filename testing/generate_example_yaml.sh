#!/bin/bash
#
# Copyright 2017 IBM Corporation
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

# The release for which the yamls are being generated
TAG=$1
if [ -z "$TAG" ]; then
    echo "Docker image tag cannot be blank."
    exit 1
fi

#########################################################
# Replace string.
# $1=Source file
# $2=Destination file
#########################################################
replace(){
    sed '
    s/${A8_TEST_ENV}/examples/g
    s/${A8_RELEASE}/'"$TAG"'/g
    s/alpine/latest/g
    ' $1 > $2
}

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
EXAMPLESDIR=$SCRIPTDIR/../examples
TEST_SCRIPTS_DIR=$SCRIPTDIR/test-scripts

##### DOCKER ######
DOCKER_DIR=$SCRIPTDIR/docker

replace $DOCKER_DIR/bookinfo.yaml $EXAMPLESDIR/docker-bookinfo.yaml
replace $DOCKER_DIR/helloworld.yaml $EXAMPLESDIR/docker-helloworld.yaml
cp $DOCKER_DIR/controlplane.yaml $EXAMPLESDIR/docker-controlplane.yaml
# Copy the rules over to examples/
cp $TEST_SCRIPTS_DIR/helloworld-default-route-rules.json $EXAMPLESDIR/docker-helloworld-default-route-rules.json
cp $TEST_SCRIPTS_DIR/helloworld-v1-v2-route-rules.json $EXAMPLESDIR/docker-helloworld-v1-v2-route-rules.json

##### K8S ######
K8S_DIR=$SCRIPTDIR/kubernetes

replace $K8S_DIR/bookinfo.yaml $EXAMPLESDIR/k8s-bookinfo.yaml
replace $K8S_DIR/helloworld.yaml $EXAMPLESDIR/k8s-helloworld.yaml
cp $K8S_DIR/controlplane.yaml $EXAMPLESDIR/k8s-controlplane.yaml
# Copy the rules over to examples/
for f in $TEST_SCRIPTS_DIR/{helloworld,bookinfo}*.yaml; do \
    cp $f $EXAMPLESDIR/k8s-`basename $f`; \
done

##### Bluemix cfg file #####
replace $SCRIPTDIR/bluemix.cfg $EXAMPLESDIR/bluemix.cfg
