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

rfile="registry.yaml"
cfile="controller.yaml"
lgfile="logserver.yaml"

if [ "$1" == "start" ]; then
    echo "Starting redis storage"
    kubectl create -f $SCRIPTDIR/$rdsfile
    echo "Starting logging service (ELK)"
    kubectl create -f $SCRIPTDIR/$lgfile
    echo "Starting multi-tenant service registry"
    kubectl create -f $SCRIPTDIR/$rfile
    echo "Starting multi-tenant controller"
    kubectl create -f $SCRIPTDIR/$cfile
    echo "Waiting for controller to initialize..."
    sleep 5
    REGISTRY_URL=$(kubectl get svc/registry --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    CONTROLLER_URL=$(kubectl get svc/controller --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})

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

    echo "Please assign a public IP to your controller"
elif [ "$1" == "stop" ]; then
    echo "Stopping control plane services.."
    kubectl delete -f $SCRIPTDIR/$cfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$rfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$lgfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$rdsfile
else
    echo "usage: $0 start|stop"
    exit 1
fi
