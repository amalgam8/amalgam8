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

if [ "$1" == "start" ]; then
    echo "starting Control plane components (ELK stack, registry, and controller)"
    docker-compose -f $SCRIPTDIR/controlplane.yaml up -d
    echo "waiting for the cluster to initialize.."

    sleep 5

    REGISTRY_URL=$(docker inspect -f '{{.NetworkSettings.IPAddress}}' registry ):8080
    CONTROLLER_URL=localhost:31200

    # Wait for controller route to set up
    echo "Waiting for controller route to set up"
    attempt=0
    while true; do
        code=$(curl -w "%{http_code}" "${CONTROLLER_URL}/health" -o /dev/null)
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

    # Wait for registry route to set up
    echo "Waiting for registry route to set up"
    attempt=0
    while true; do
        code=$(curl -w "%{http_code}" "${REGISTRY_URL}/uptime" -o /dev/null)
        if [ "$code" = "200" ]; then
            echo "Registry route is set to '$REGISTRY_URL'"
            break
        fi

        attempt=$((attempt + 1))
        if [ "$attempt" -gt 10 ]; then
            echo "Timeout waiting for registry route: /uptime returned HTTP ${code}"
            echo "Deploying the controlplane has failed"
            exit 1
        fi
        sleep 10s
    done


    echo "Setting up a new tenant named 'local'"
    read -d '' tenant << EOF
{
    "load_balance": "round_robin"
}
EOF
    echo $tenant | curl -H "Content-Type: application/json" -H "Authorization: Bearer local" -d @- "http://${CONTROLLER_URL}/v1/tenants"
elif [ "$1" == "stop" ]; then
    echo "Stopping control plane services..."
    docker-compose -f $SCRIPTDIR/controlplane.yaml kill
    docker-compose -f $SCRIPTDIR/controlplane.yaml rm -f
else
    echo "usage: $0 start|stop"
    exit 1
fi
