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


#gunicorn -D -w 1 -b 0.0.0.0:10081 --reload details:app
#gunicorn -D -w 1 -b 0.0.0.0:10082 --reload reviews:app
#gunicorn -w 1 -b 0.0.0.0:19080 --reload --access-logfile prod.log --error-logfile prod.log productpage:app >>prod.log 2>&1 &
kubectl create -f bookinfo.yaml
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED to create bookinfo app.\n***********\n"
    exit 1
fi
# echo "waiting for a few seconds for pods to get created.."
# sleep 10
# echo "setting version routing rules for productpage, reviews, ratings, and details"
# AC=$(kubectl get svc/controller --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
# curl -X PUT ${AC}/v1/tenants/local/versions/productpage -d '{"default": "v1"}' -H "Content-Type: application/json" 
# #curl -X PUT ${AC}/v1/tenants/local/versions/reviews -d '{"default": "v1", "selectors": "{v2={weight=0.3},v3={weight=0.2}}" }' -H "Content-Type: application/json" 
# curl -X PUT ${AC}/v1/tenants/local/versions/reviews -d '{"default": "v1", "selectors": "{v2={user=\"frankb\"},v3={user=\"shriram\"}}" }' -H "Content-Type: application/json" 
# curl -X PUT ${AC}/v1/tenants/local/versions/ratings -d '{"default": "v1"}' -H "Content-Type: application/json" 
# curl -X PUT ${AC}/v1/tenants/local/versions/details -d '{"default": "v1"}' -H "Content-Type: application/json" 
