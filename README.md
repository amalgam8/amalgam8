# Amalgam8 examples

Sample microservice-based applications and local sandbox environment for Amalgam8

## Overview

This project includes a number of Amalgam8 sample programs, and a preconfigured environment to allow
you to easily run, build, and experiment with the provided samples, or use your own ideas.

The following samples are available for Amalgam8:

* **Helloworld** is a single microservice app that demonstrates how to route traffic to different versions of an instance in Amalgam8
* **Bookinfo** is a multiple microservice app used to demonstrate and experiment with several Amalgam8 features

There is also an end-to-end [Test & Deploy demo](https://github.com/amalgam8/examples/blob/master/demo-script.md) that you can try.

The root directory includes a Vagrant file that provides an environment with everything installed and ready to build/run the samples:

* [Go](http://golang.org/)
* [Docker](http://www.docker.com/)
* [Kubernetes](http://kubernetes.io/)
* [Amalgam8 CLI](https://github.com/amalgam8/controller/tree/master/cli)

To get started, follow the instructions to set up the sandbox environment, and then run any or all of the samples
as described below.

## Setup the sandbox environment

1. Download and install [Git](https://git-scm.com/)
2. Download and install [Vagrant](https://www.vagrantup.com/)
3. On Windows only, download and install [VirtualBox](https://www.virtualbox.org/). Run all commands in the steps from VirtualBox.
4. Run the following commands to clone amalgam8 and to start the vagrant environment.

  ```
    git clone git@github.com:amalgam8/examples.git
    git clone git@github.com:amalgam8/registry.git
    git clone git@github.com:amalgam8/controller.git
    git clone git@github.com:amalgam8/sidecar.git

    cd examples
    vagrant up
    vagrant ssh
  ```

  Note: If you stopped a previous Vagrant VM and restarted it, Kubernetes might not run correctly. If you have problems, try uninstalling Kubernetes by running the following commands: 
  
  ```
    cd $GOPATH/src/github.com/amalgam8/examples
    run ./uninstall-kubernetes.sh
    run ./install-kubernetes.sh (You can ignore the Permission denied message, if you see one).*
  ```
  
  Then reinstall Kubernetes, by running the following command. If a *Permission denied* message appears, you can safely ignore it.
  
  ```
    run ./install-kubernetes.sh
  ```
  
### Running the controlplane services

5. Start the local control plane services (registry and controller) by running the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/controlplane
    ./run-controlplane.sh compile
    ./run-controlplane.sh start
  ```

6. Run the following commands to confirm whether the registry and controller services are running:

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

7. (a) To list your registered services, use the following command format:

  ```
    $ export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY
    $ curl -X GET -H "Authorization: Bearer ${TOKEN}" http://10.0.0.230:5080/api/v1/services | jq .
    {
      "services": []
    }
  ```

7. (b) To view your tenant entry in the controller, use the following command format:

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

8. To start the API gateway, run the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/gateway
    kubectl create -f gateway.yaml
  ```

  Usually, the API gateway is mapped to a DNS route. However, in our local standalone environment, you can access it by using
  the fixed IP address and port (192.168.33.33:32000), which was preconfigured for the sandbox environment.

9. Confirm that the API gateway is running by running the following command:

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

## Running the samples

10. Follow the instructions in the README for the sample that you want to use.
  (a) *helloworld* sample

  See https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md

  (b) *bookinfo* sample

  See https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md

## Shutting down

11. When you are finished, to shut down the gateway and control plane servers, run the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples
    kubectl delete -f gateway/gateway.yaml
    controlplane/run-controlplane.sh stop
  ```
