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

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

ctrlplanefile="controlplane.yaml"

if [ "$1" == "start" ]; then

    echo "Starting control plane"
    kubectl create -f $SCRIPTDIR/$ctrlplanefile

    echo "Waiting for control plane to initialize..."

    sleep 10
    CONTROLLER_URL=http://localhost:31200

    # Wait for controller route to set up
    echo "Waiting for controller route to set up"
    attempt=0
    while true; do
        code=$(curl -w "%{http_code}" --max-time 10 "${CONTROLLER_URL}/health" -o /dev/null)
        if [ "$code" = "200" ]; then
            echo "Controller route is set to '$CONTROLLER_URL'"
            break
        fi

        attempt=$((attempt + 1))
        if [ "$attempt" -gt 10 ]; then
            echo "Timeout waiting for controller route: /health returned HTTP ${code}"
            echo "Deploying the controlplane has failed"
            exit 1
        fi
        sleep 10s
    done

elif [ "$1" == "stop" ]; then
    echo "Stopping control plane services..."
    kubectl delete -f $SCRIPTDIR/$ctrlplanefile
    sleep 3
    # kubectl delete -f $SCRIPTDIR/$cfile
    # sleep 3
    # kubectl delete -f $SCRIPTDIR/$rdsfile
else
    echo "usage: $0 start|stop"
    exit 1
fi
