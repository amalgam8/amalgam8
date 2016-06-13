# Amalgam8 Developer Instructions

The easiest way to set up an environment where you can compile and experiment with Amalgam8 source code
is by using the same vagrant VM that is used for the kick-the-tires demos in
the [examples](https://github.com/amalgam8/examples) project, only in this case you need to pull more
git repos before doing the "vagrant up".
Alternatively you can set up the required prereqs, described in the 
[Vagrantfile](https://github.com/amalgam8/examples/blob/master/Vagrantfile), yourself,
and then run the samples on your own machine of choice.

To get started using the provided Vagrantfile, proceed as follows:

```bash
git clone git@github.com:amalgam8/examples.git
git clone git@github.com:amalgam8/registry.git
git clone git@github.com:amalgam8/controller.git
git clone git@github.com:amalgam8/sidecar.git

cd examples
vagrant up
vagrant ssh
```

*Note:* If you stopped a previous Vagrant VM and restarted it, Kubernetes might not run correctly. If you have problems, try uninstalling Kubernetes by running the following commands: 
  
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

An [API Gateway](http://microservices.io/patterns/apigateway.html) provides
a single user-facing entry point for a microservices-based application.
You can control the Amalgam8 gateway for different purposes, such as
version routing, red/black deployments, canary testing, resiliency
testing, and so on.

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

  Note: You only need one gateway per tenant. A single gateway can front
  more than one application under the tenant at the same time, so long as
  they don't implement any conflicting microservices.

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
