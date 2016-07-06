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

AR=$(kubectl get svc/registry --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
AC=localhost:31200
KA=$(kubectl get svc/kafka --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
echo "Setting up a new tenant named 'local'"
read -d '' tenant << EOF
{
    "req_tracking_header" : "X-Request-ID",
    "credentials": {
        "kafka": {
            "brokers": ["${KA}"],
            "sasl": false
        },
        "registry": {
            "url": "http://${AR}",
            "token": "local"
        }
    }
}
EOF
echo $tenant | curl -H "Content-Type: application/json" -H "Authorization: local" -d @- "http://${AC}/v1/tenants"
