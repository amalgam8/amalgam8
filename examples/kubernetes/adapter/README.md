This directory contains scripts to run the helloworld sample using the Kubernetes registry adapter.
The helloworld instances do not need to register/heartbeat their services because the helloworld service 
is defined in Kubernetes, and the registry adapter then automatically reflects the K8s services in the A8 Registry.

```bash
cd amalgam8/examples/kubernetes/adapter
./run-controlplane-local-k8s.sh start
kubectl create -f ../gateway.yaml
kubectl create -f helloworld-svc.yaml
```

After that, you can run the usual a8ctl and curl commands and see the same
behavior as in the helloworld demo. You can also query the Kubernetes API
service to obtain the list of helloworld service instances.

```bash
curl http://localhost:8080/api/v1/namespaces/local/endpoints/helloworld
```
