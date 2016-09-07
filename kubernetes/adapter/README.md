This directory contains scripts to run the helloworld sample using the Kubernetes registry adapter.
The helloworld instances do not need to register/heartbeat their services because the helloworld service 
is defined in Kubernetes, and the registry adapter then automatically reflects the K8s services in the A8 Registry.

First make sure you have a built registry Docker image that includes the K8s adapter code, and then run the following:

```
export A8_REGISTRY_TOKEN="local"
export A8_CONTROLLER_TENANT_ID="local" # export A8_CONTROLLER_TOKEN="local"

cd kubernetes/adapter
./run-controlplane-local-k8s.sh start
kubectl create -f ../gateway.yaml
kubectl create -f helloworld-svc.yaml
```

After that, you can run the usual a8ctl and curl commands and see the same behavior as in the helloworld demo.

```
# a8ctl service-list
# a8ctl route-set helloworld --default v1 --selector 'v2(weight=0.5)'
#   

# curl http://localhost:32000/helloworld/hello
# curl http://localhost:8080/api/v1/namespaces/local/endpoints/helloworld
# curl -X GET -H "Authorization: Bearer local" http://localhost:31300/api/v1/services/helloworld | jq .
# curl http://localhost:31200/v1/tenants/local/versions/helloworld | jq .

# kubectl delete -f helloworld-svc.yaml
# kubectl delete -f gateway.yaml
```
