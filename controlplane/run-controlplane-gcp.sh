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
mhfile="messagehub.yaml"
lgfile="logserver.yaml"

if [ "$1" == "compile" ]; then
    $SCRIPTDIR/build-sidecar.sh
    if [ $? -ne 0 ]; then
        echo "Sidecar failed to compile"
        exit 1
    fi
    $SCRIPTDIR/build-registry.sh
    if [ $? -ne 0 ]; then
        echo "Registry failed to compile"
        exit 1
    fi
    $SCRIPTDIR/build-controller.sh
    if [ $? -ne 0 ]; then
        echo "Controller failed to compile"
        exit 1
    fi
    exit
elif [ "$1" == "start" ]; then
    echo "Starting integration bus (kafka)"
    kubectl create -f $SCRIPTDIR/$mhfile
    echo "Starting logging service (ELK)"
    kubectl create -f $SCRIPTDIR/$lgfile
    echo "Starting multi-tenant service registry"
    kubectl create -f $SCRIPTDIR/$rfile
    echo "Starting multi-tenant controller"
    kubectl create -f $SCRIPTDIR/$cfile
    echo "Waiting for controller to initialize..."
    sleep 60
    AR=$(kubectl get svc/registry --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    AC=$(kubectl get svc/controller --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    KA=$(kubectl get svc/kafka --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    echo "Setting up a new tenant named 'local'"
    read -d '' tenant << EOF
{
    "id": "local",
    "token": "local",
    "req_tracking_header" : "X-Request-ID",
    "credentials": {
        "kafka": {
            "brokers": ["${KA}"],
            "sasl": false
        },
        "registry": {
            "url": "http://${AR}",
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY"
        }
    }
}
EOF
    export TENANT_REG=$tenant
    echo "Please assign a public IP to your controller and then issue the following curl command"
    echo 'echo $TENANT_REG|curl -H "Content-Type: application/json" -d @- http://ControllerExternalIP:31200/v1/tenants'

    echo $tenant | curl -H "Content-Type: application/json" -d @- "http://${AC}/v1/tenants"
elif [ "$1" == "stop" ]; then
    echo "Stopping control plane services.."
    kubectl delete -f $SCRIPTDIR/$cfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$rfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$lgfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$mhfile
else
    echo "usage: $0 compile|start|stop"
    exit 1
fi
