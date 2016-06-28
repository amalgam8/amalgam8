This directory contains scripts to run the helloworld sample using the Kubernetes registry adapter.
The helloworld instances do not need to register/heartbeat their services because the helloworld service 
is defined in Kubernetes, and the registry adapter then automatically reflects the K8s services in the A8 Registry.

First make sure you have a built Docker image that includes the K8s adapter code, and then run the following:

```
export A8_REGISTRY_TOKEN="klocal"
export A8_CONTROLLER_TENANT_ID="klocal" # export A8_CONTROLLER_TOKEN="klocal"

./run-controlplane-local-k8s.sh start
kubectl create -f gateway.yaml
kubectl create -f helloworld-svc.yaml
```

After that, you can run the usual a8ctl and curl commands and see the same behavior as in the helloworld demo.

```
# a8ctl service-list
# a8ctl route-set helloworld --default v1 --selector 'v2(weight=0.5)'
# a8ctl route-list

# curl http://localhost:8080/api/v1/namespaces/klocal/endpoints/helloworld
# curl -X GET -H "Authorization: Bearer klocal" http://localhost:31300/api/v1/services/helloworld | jq .
# curl http://localhost:31200/v1/tenants/klocal/versions/helloworld | jq .

# kubectl delete -f helloworld-svc.yaml
# kubectl delete -f gateway.yaml
```
