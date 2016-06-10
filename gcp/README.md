Running Amalgam8 on Google Cloud Platform
-----

* Setup [Google Cloud SDK](https://cloud.google.com/sdk/) on your machine

```bash
./run-controlplane.sh start
```

* You need to assign an external IP to your controller so that you can
  register tenants and communicate with it.

* One external IP is enough to share amongst all the components (logserver,
  controller, etc.)

* Visualizing your deployment with Weave scope

```bash
kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false
```

Once weavescope is up and running, you can view the weavescope dashboard  on your local host using the following commands
  
```bash
kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath={.items..metadata.name}) 4040
```

You can open http://localhost:4040 on your browser to access the Scope UI
securely and visualize your K8S deployment.

* The rest of the sample apps work as is on Google Cloud Platform, as they
  do on a local vagrant deployment with Kubernetes.
  
