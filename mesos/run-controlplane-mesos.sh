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
    echo "starting Marathon/Mesos cluster with Kafka + ELK stack"
    docker-compose up -d
    echo "waiting for the cluster to initialize.."
    sleep 60
    echo "Starting multi-tenant service registry"
    cat registry.json|curl -X POST -H "Content-Type: application/json" http://192.168.33.33:8080/v2/apps -d@-
    echo "Starting multi-tenant controller"
    cat controller.json|curl -X POST -H "Content-Type: application/json" http://192.168.33.33:8080/v2/apps -d@-
    echo "Waiting for controller to initialize..."
    sleep 20
    AR="192.168.33.33:31300"
    AC="192.168.33.33:31200"
    KA="192.168.33.33:9092"
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
    echo $tenant | curl -H "Content-Type: application/json" -d @- "http://${AC}/v1/tenants"
elif [ "$1" == "stop" ]; then
    echo "Stopping control plane services..."
    curl -X DELETE -H "Content-Type: application/json" http://192.168.33.33:8080/v2/apps/a8-controller
    curl -X DELETE -H "Content-Type: application/json" http://192.168.33.33:8080/v2/apps/a8-registry
    sleep 10
    docker-compose kill
    docker-compose rm -f
else
    echo "usage: $0 start|stop"
    exit 1
fi
