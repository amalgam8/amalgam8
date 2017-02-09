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

# integration(default) or examples
A8_TEST_SUITE=$1

A8_RELEASE=$2
# Remove v from version
A8_RELEASE=$(echo $A8_RELEASE | sed "s/v//")


if [ -z "$A8_TEST_DOCKER" ]; then
    A8_TEST_DOCKER="true"
fi

if [ -z "$A8_TEST_K8S" ]; then
    A8_TEST_K8S="true"
fi

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$A8_TEST_DOCKER" == "true" ]; then
    $SCRIPTDIR/docker/test-docker.sh $A8_TEST_SUITE $A8_RELEASE
fi

if [ "$A8_TEST_K8S" == "true" ]; then
    $SCRIPTDIR/kubernetes/test-kubernetes.sh $A8_TEST_SUITE $A8_RELEASE
fi
