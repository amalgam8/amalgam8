# Amalgam8 examples

Sample microservice-based applications and local sandbox environment for Amalgam8

## Table of Contents

* [Overview](#overview)
* [Amalgam8 on Kubernetes (local)](#local-k8s)
* [Amalgam8 on Marathon/Mesos (local)](#local-marathon)
* [Amalgam8 on Google Compute Cloud](#gcp)

## Overview <a id="overview"></a>

This project includes a number of Amalgam8 sample programs, and a preconfigured environment to allow
you to easily run, build, and experiment with the provided samples.

The following samples are available for Amalgam8:

* **Helloworld** is a single microservice app that demonstrates how to route traffic to different versions of the same microservice
* **Bookinfo** is a multiple microservice app used to demonstrate and experiment with several Amalgam8 features

There is an end-to-end
[test & deploy demo](https://github.com/amalgam8/examples/blob/master/demo-script.md)
that you can try out on your local vagrant box.

In addition, the scripts are generic enough such that you can easily deploy
them to any kubernetes based environment such as OpenShift, Google Cloud
Platform, Azure, etc. See the last section of this README for details on
how to deploy Amalgam8 on Google Cloud Platform.

## Amalgam8 with Kubernetes - local environment <a id="local-k8s"></a>

The repository's root directory includes a Vagrant file that provides an environment with everything installed and ready to build/run the samples:

* [Go](http://golang.org/)
* [Docker](http://www.docker.com/)
* [Kubernetes](http://kubernetes.io/)
* [Amalgam8 CLI](https://github.com/amalgam8/controller/tree/master/cli)

To get started, install a recent version of Vagrant and follow the steps below.

1. Clone the Amalgam8 repos and start the vagrant environment.

  ```
    git clone git@github.com:amalgam8/examples.git
    git clone git@github.com:amalgam8/registry.git
    git clone git@github.com:amalgam8/controller.git
    git clone git@github.com:amalgam8/sidecar.git

    cd examples
    vagrant up
    vagrant ssh
  ```

  *Note*: If you stopped a previous Vagrant VM and restarted it, Kubernetes might not run correctly. If you have problems, try uninstalling Kubernetes by running the following commands: 
  
  ```
    cd $GOPATH/src/github.com/amalgam8/examples
    sudo ./uninstall-kubernetes.sh
  ```

  Then re-install Kubernetes, by running the following command:

  ```
    sudo ./install-kubernetes.sh
  ```

### Running the controlplane services

2. Start the local control plane services (registry and controller) by running the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/controlplane
    ./run-controlplane-local.sh compile
    ./run-controlplane-local.sh start
  ```

3. Run the following commands to confirm whether the registry and controller services are running:

  ```
    kubectl get svc
  ```

  If the registry and controller services are running, the output will resemble the following example:

  ```
    NAME               CLUSTER_IP   EXTERNAL_IP   PORT(S)    SELECTOR                AGE
    kubernetes         10.0.0.1     <none>        443/TCP    <none>                  40d
    registry           10.0.0.230    <none>        5080/TCP   name=registry           1m
    controller         10.0.0.240    <none>        6379/TCP   name=controller         1m
  ```

  You can reach the registry at 10.0.0.230:5080, and the controller at
  10.0.0.240:6379. You can also reach the controller from
  outside the vagrant box at 192.168.33.33:31200. You can use cURL if
  you want to see them working.

4. (a) To list your registered services, use the following command format:

  ```
    $ export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY
    $ curl -X GET -H "Authorization: Bearer ${TOKEN}" http://10.0.0.230:5080/api/v1/services | jq .
    {
      "services": []
    }
  ```

5. (b) To view your tenant entry in the controller, use the following command format:

  ```
    curl http://10.0.0.240:6379/v1/tenants/local | jq .
    {
      "filters": {
        "versions": [],
        "rules": []
      },
      "port": 6379,
      "load_balance": "round_robin",
      "credentials": {
        "registry": {
          "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY",
          "url": "http://10.0.0.230:5080"
        },
        "message_hub": {
          "sasl": false,
          "password": "",
          "user": "",
          "kafka_broker_sasl": [
            "10.0.0.200:9092"
          ],
          "kafka_rest_url": "",
          "kafka_admin_url": "",
          "api_key": ""
        }
      },
      "id": "local"
    }
  ```

### Running the API Gateway

  An [API Gateway](http://microservices.io/patterns/apigateway.html) provides a single user-facing entry point for a microservices-based application.
  You can control the Amalgam8 gateway for different purposes, such as active deploy, resiliency testing, and so on.

6. To start the API gateway, run the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/gateway
    kubectl create -f gateway.yaml
  ```

  Usually, the API gateway is mapped to a DNS route. However, in our local standalone environment, you can access it by using
  the fixed IP address and port (192.168.33.33:32000), which was preconfigured for the sandbox environment.

7. Confirm that the API gateway is running by running the following command:

  ```
    curl 192.168.33.33:32000/
  ```

  If the gateway is running, the output will resemble the following example:

  ```
    <!DOCTYPE html>
    <html>
    <head>
    <title>Welcome to nginx!</title>
    <style>
        body {
            width: 35em;
            margin: 0 auto;
            font-family: Tahoma, Verdana, Arial, sans-serif;
        }
    </style>
    </head>
    <body>
    <h1>Welcome to nginx!</h1>
    <p>If you see this page, the nginx web server is successfully installed and
    working. Further configuration is required.</p>

    <p>For online documentation and support please refer to
    <a href="http://nginx.org/">nginx.org</a>.<br/>
    Commercial support is available at
    <a href="http://nginx.com/">nginx.com</a>.</p>

    <p><em>Thank you for using nginx.</em></p>
    </body>
    </html>
  ```

  Note: A single gateway can front more than one sample app at the same time, so long as they don't implement any conflicting microservices.

  Now that the control plane services and gateway are running, you can run the samples.

### Running the samples

8. Follow the instructions in the README for the sample that you want to use.
  (a) *helloworld* sample

  See https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md

  (b) *bookinfo* sample

  See https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md

### Shutting down

9. When you are finished, to shut down the gateway and control plane servers, run the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples
    kubectl delete -f gateway/gateway.yaml
    controlplane/run-controlplane-local.sh stop
  ```

## Amalgam8 with Marathon/Mesos - local environment <a id="local-marathon"></a>

This section assumes that the IP address of your mesos slave where all the
apps will be running is 192.168.33.33.

1. The `run-controlplane-mesos.sh` script in the `mesos` folder sets up a
   local marathon/mesos cluster (based on Holiday Check's
   [mesos-in-the-box](https://github.com/holidaycheck/mesos-in-the-box))  and launches the controller and the
   registry as apps in the marathon framework.
   
```bash
cd mesos
./run-controlplane-mesos.sh start
```

Make sure that the Marathon dashboard is accessible at http://192.168.33.33:8080 and the Mesos dashboard at http://192.168.33.33:5050

Verify that the controller is up and running via the Marathon dashboard.

2. Launch the API Gateway

```bash
cat gateway.json|curl -X POST -H "Content-Type: application/json" http://192.168.33.33:8080/v2/apps -d@-
```

Verify that the gateway is reacheable by accessing http://192.168.33.33:32000

3. Launch the Bookinfo application

```bash
cat bookinfo.json|curl -X POST -H "Content-Type: application/json" http://192.168.33.33:8080/v2/groups -d@-
```

Verify that the application group has been successfully launched via the marathon dashboard.

4. You can now use the `a8ctl` command line tool to set the default
versions for various services in the Bookinfo app, do version-based
routing, resilience testing etc. For more details, refer to the
[test & deploy demo](https://github.com/amalgam8/examples/blob/master/demo-script.md)

## Amalgam8 on Google Cloud Platform <a id="gcp"></a>

1. Setup [Google Cloud SDK](https://cloud.google.com/sdk/) on your machine

2. Setup a cluster of 3 nodes

3. Launch the control plane services

```bash
controlplane/run-controlplane-gcp.sh start
```

4. Locate the node where the controller is running and assign an
  external IP to the node if needed

5. Initialize the first tenant. The `run-controlplane-gcp.sh` script stores
   the JSON payload to initialize the tenant in the `TENANT_REG` environment variable.

```bash
echo $TENANT_REG|curl -H "Content-Type: application/json" -d @- http://ControllerExternalIP:31200/v1/tenants'
```

6. Deploy the API gateway

```bash
kubectl create -f gateway/gateway.yaml
```

Obtain the public IP of the node where the gateway is running. This will be
the be IP at which the sample app will be accessible.

7. You can now deploy the sample apps as described in "Running the sample
  apps" section above. Remember to replace the IP address `192.168.33.33`
  with the public IP address of the node where the gateway service is
  running on the Google Cloud Platform.

8. Visualizing your deployment with Weave Scope

```bash
kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false
```

Once weavescope is up and running, you can view the weavescope dashboard
on your local host using the following commands
  
```bash
kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath={.items..metadata.name}) 4040
```
  
You can open http://localhost:4040 on your browser to access the Scope UI.
