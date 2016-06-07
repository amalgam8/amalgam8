#!/bin/bash

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

docker build -t hello:vx $SCRIPTDIR
kubectl create -f $SCRIPTDIR/helloworld.yaml

#AC=`kubectl get svc controller --no-headers|awk '{print $2}'`
#curl -H "Authorization: 12345" -H "Content-Type: application/json" -X PUT ${AC}:6379/v1/tenants/local/versions/helloworld -d '{"default": "v1"}'
