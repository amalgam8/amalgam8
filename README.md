# Amalgam8 Examples

Sample microservice-based applications and local sandbox environment for Amalgam8.
An overview of Amalgam8 can be found at www.amalgam8.io.

## Overview <a id="overview"></a>

This project includes a number of Amalgam8 sample programs, scripts and a preconfigured environment to allow
you to easily run, build, and experiment with the provided samples, in several environments.
In addition, the scripts are generic enough that you can easily deploy
the samples to other environments as well.

The following samples are available for Amalgam8:

* [Helloworld](https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md) is a single microservice app that demonstrates how to route traffic to different versions of the same microservice
* [Bookinfo](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md) is a multiple microservice app used to demonstrate and experiment with several Amalgam8 features

The repository's root directory includes a Vagrant file that provides an environment with everything 
needed to run, and build, the samples
([Go](http://golang.org/), [Docker](http://www.docker.com/), [Kubernetes](http://kubernetes.io/),
[Amalgam8 CLI](https://github.com/amalgam8/controller/tree/master/cli), [Gremlin SDK](https://github.com/ResilienceTesting/gremlinsdk-python))
already installed. Depending on the runtime environement you want to try, using it may be the easiest way to get Amalgam8 up and running.

To run the demos, proceed to the instructions corresponding to the environment that you want to use:

* Localhost Deployment
    * [Docker](#local-docker)
    * [Kubernetes](#local-k8s)
    * [Marathon/Mesos](#local-marathon)
* Cloud Deployment
    * [IBM Bluemix](#bluemix)
    * [Google Compute Cloud](#gcp)

If you'd like to also be able to change and compile the code, or build the images,
refer the [Developer Instructions](https://github.com/amalgam8/examples/blob/master/development.md).


## Amalgam8 with Docker - local environment <a id="local-docker"></a>

To run in a local docker environemnt, you can either use the vagrant sandbox
or you can simply install [Docker Toolbox](https://www.docker.com/products/docker-toolbox),
and the [Amalgam8 CLI](https://pypi.python.org/pypi/a8ctl) on your machine.

1. Clone the Amalgam8 examples repo and then start the vagrant environment (or install and setup the equivalent dependencies manually)

    ```bash
    git clone git@github.com:amalgam8/examples.git

    cd examples
    vagrant up
    vagrant ssh

    cd $GOPATH/src/github.com/amalgam8/examples
    ```

1. Start the multi-tenant control plane and a tenant

    Start the control plane services (registry and controller) by running the
    following command:

    ```bash
    docker/run-controlplane-docker.sh start
    ```

    The above command also creates a tenant named "local" in the
    control plane. 

1. Before we start using the `a8ctl` command line utility, we need to point
   it to the address of the controller and the registry.

    * If you are running the docker setup using the Vagrant file in the
    `examples` folder, or you are running Docker locally.

    ```bash
    export A8_CONTROLLER_URL=http://localhost:31200
    export A8_REGISTRY_URL=http://localhost:31300
    ```

    * If you are running Docker on Mac/Windows using the Docker Toolbox (__not
    the Docker for Mac - Beta__), then set the environment variables to the IP
    address of the VM created by Docker Machine.

    Assuming you have only one Docker Machine running on your system, the
    following commands will setup the appropriate environment variables:

    ```bash
    export A8_CONTROLLER_URL=`docker-machine ip`
    export A8_REGISTRY_URL=`docker-machine ip`
    ```

    You get the idea. Just setup these two environment variables appropriately
    and you should be good to go.

1.  Confirm everything is working with the following command:

    ```bash
    a8ctl service-list
    ```

    The command shouldn't return any services, since we haven't started any yet, 
    but if it returns the following empty table, the control plane services (and CLI) are working as expected:

    ```
    +---------+-----------+
    | Service | Instances |
    +---------+-----------+
    +---------+-----------+
    ```

1. Deploy the API gateway

    Every tenant application should have
    an [API Gateway](http://microservices.io/patterns/apigateway.html) that
    provides a single user-facing entry point for a microservices-based
    application.  You can control the Amalgam8 gateway for different purposes,
    such as version routing, red/black deployments, canary testing, resiliency
    testing, and so on. The Amalgam8 gateway is a simple lightweight Nginx
    server that is controlled by the control plane.


    To start the API gateway, run the following command:

    ```bash
    docker-compose -f docker/gateway.yaml up -d
    ```

    Usually, the API gateway is mapped to a DNS route. However, in our local
    standalone environment, you can access it at port 32000 on localhost.
    If you are using docker directly, then the gateway should be
    accessible at http://localhost:32000 or http://dockermachineip:32000.

1. Confirm that the API gateway is running by accessing the URL from your
    browser. If all is well, you should see a simple **Welcome to nginx!**
    page in your browser.

    **Note:** You only need one gateway per tenant. A single gateway can front more
    than one application under the tenant at the same time, so long as they
    don't implement any conflicting microservices.

1. Following instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:

    ```bash
    docker-compose -f docker/helloworld.yaml up -d
    ```
        
    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md

    * To shutdown the helloworld instances, run the following commands:
   
    ```bash
    docker-compose -f docker/helloworld.yaml kill
    docker-compose -f docker/helloworld.yaml rm -f
    ```

    (b) **bookinfo** sample

    * Start the bookinfo application:
    
    ```bash
    docker-compose -f docker/bookinfo.yaml up -d
    ```

    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md

    * To shutdown the bookinfo instances, run the following commands:
    
    ```
    docker-compose -f docker/bookinfo.yaml kill
    docker-compose -f docker/bookinfo.yaml rm -f
    ```

    When you are finished, shut down the gateway and control plane servers by running the following commands:

    ```
    docker-compose -f docker/gateway.yaml kill
    docker-compose -f docker/gateway.yaml rm -f
    docker/run-controlplane-docker.sh stop
    ```

## Amalgam8 with Kubernetes - local environment <a id="local-k8s"></a>

1. Clone the Amalgam8 examples repo and then start the vagrant environment (or install and setup the equivalent dependencies manually)

    ```bash
    git clone git@github.com:amalgam8/examples.git
    
    cd examples
    vagrant up
    vagrant ssh

    cd $GOPATH/src/github.com/amalgam8/examples
    export A8_CONTROLLER_URL=http://localhost:31200
    ```
    
    *Note:* If you stopped a previous Vagrant VM and restarted it, Kubernetes might not run correctly. If you have problems, try uninstalling Kubernetes by running the following commands: 
      
    ```bash
    sudo kubernetes/uninstall-kubernetes.sh
    ```
    
    Then re-install Kubernetes, by running the following command:
    
    ```bash
    sudo kubernetes/install-kubernetes.sh
    ```
    
    **Note:** if you do reinstall kubernetes, wait until it has initialized (i.e., kubectl commands are working) before proceeding to the next step.

1. Start the local control plane services (registry and controller) by running the following commands:

    ```bash
    kubernetes/run-controlplane-local-k8s.sh start
    ```

1. Run the following command to confirm the control plane is working:

    ```bash
    a8ctl service-list
    ```

    The command shouldn't return any services, since we haven't started any yet, 
    but if it returns the following empty table, the control plane servers (and CLI) are working as expected:
    
    ```
    +---------+-----------+
    | Service | Instances |
    +---------+-----------+
    +---------+-----------+
    ```
    
    You can also access the registry at http://localhost:31300 from the host machine
    (outside the vagrant box), and the controller at http://localhost:31200.
    To access the control plane details of tenant *local*, access
    http://localhost:31200/v1/tenants/local from your browser.

1. Run the [API Gateway](http://microservices.io/patterns/apigateway.html) with the following commands:

    ```bash
    kubectl create -f kubernetes/gateway.yaml
    ```
    
    Usually, the API gateway is mapped to a DNS route. However, in our local
    standalone environment, you can access it at port 32000 on localhost.

1. Confirm that the API gateway is running by accessing the
    http://localhost:32000 from your browser. If all is well, you should
    see a simple **Welcome to nginx!** page in your browser.

    **Note:** You only need one gateway per tenant. A single gateway can front more
    than one application under the tenant at the same time, so long as they
    don't implement any conflicting microservices.

1. Following instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:
    
        ```bash
        kubectl create -f kubernetes/helloworld.yaml
        ```
        
    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md
 
    * To shutdown the helloworld instances, run the following command:
    
        ```bash
        kubectl delete -f kubernetes/helloworld.yaml
        ```

    (b) **bookinfo** sample

    * Start the bookinfo application:
    
        ```bash
        kubectl create -f kubernetes/bookinfo.yaml
        ```

    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md
    
    * To shutdown the bookinfo instances, run the following command:
    
        ```bash
        kubectl delete -f kubernetes/bookinfo.yaml
        ```

1. When you are finished, shut down the gateway and control plane servers by running the following commands:

    ```bash
    kubectl delete -f kubernetes/gateway.yaml
    kubernetes/run-controlplane-local-k8s.sh stop
    ```

## Amalgam8 with Marathon/Mesos - local environment <a id="local-marathon"></a>

1. Clone the Amalgam8 examples repo and then start the vagrant environment (or install and setup the equivalent dependencies manually)

    ```bash
    git clone git@github.com:amalgam8/examples.git
    
    cd examples
    vagrant up
    vagrant ssh

    cd $GOPATH/src/github.com/amalgam8/examples
    ```

1. The `run-controlplane-marathon.sh` script in the `marathon` folder sets up a
   local marathon/mesos cluster (based on Holiday Check's
   [mesos-in-the-box](https://github.com/holidaycheck/mesos-in-the-box))  and launches the controller and the
   registry as apps in the marathon framework.
   
    ```bash
    marathon/run-controlplane-marathon.sh start
    ```

    This section assumes that your mesos slave, where all
    the apps will be running, is on localhost.

    Make sure that the Marathon dashboard is accessible at http://localhost:8080 and the Mesos dashboard at http://localhost:5050

    Verify that the controller is up and running via the Marathon dashboard.

1. Launch the API Gateway
    
    ```bash
    cat marathon/gateway.json|curl -X POST -H "Content-Type: application/json" http://localhost:8080/v2/apps -d@-
    ```

1. Confirm that the API gateway is running by accessing the
    http://localhost:32000 from your browser. If all is well, you should
    see a simple **Welcome to nginx!** page in your browser.

    **Note:** You only need one gateway per tenant. A single gateway can front more
    than one application under the tenant at the same time, so long as they
    don't implement any conflicting microservices.

1. Following instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:

    ```bash
    cat marathon/helloworld.json| curl -X POST -H "Content-Type: application/json" http://localhost:8080/v2/groups -d@-
    ```
        
    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/helloworld/README.md

    * To shutdown the helloworld instances, run the following commands:
   
    ```bash
    curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/v2/groups/helloworld
    ```

    (b) **bookinfo** sample

    * Start the bookinfo application:
    
    ```bash
    cat marathon/bookinfo.json| curl -X POST -H "Content-Type: application/json" http://localhost:8080/v2/groups -d@-
    ```

    * Follow the instructions at https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md

    * To shutdown the bookinfo instances, run the following commands:
    
    ```bash
    curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/v2/groups/bookinfo
    ```

    When you are finished, shut down the gateway and control plane servers by running the following commands:

    ```bash
    curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/v2/apps/gateway
    curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/v2/apps/a8-controller
    curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/v2/apps/a8-registry
    ```

## Amalgam8 on IBM Bluemix <a id="bluemix"></a>

To run the [Bookinfo sample app](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md)
on Bluemix, follow the instructions below.
If you are not a bluemix user, you can register at [bluemix.net](http://bluemix.net/).

1. Download [Docker 1.10 or later](https://docs.docker.com/engine/installation/),
    [CF CLI 6.12.0 or later](https://github.com/cloudfoundry/cli/releases),
    [CF CLI IBM Containers plugin](https://console.ng.bluemix.net/docs/containers/container_cli_ov.html),
    and the [Amalgam8 CLI](https://pypi.python.org/pypi/a8ctl)
  
1. Login to Bluemix and initialize the containers environment using ```cf login``` and ```cf ic init```

1. Configure the [.bluemixrc file](bluemix/.bluemixrc) to your environment variable values
    * BLUEMIX_REGISTRY_NAMESPACE should be your Bluemix registry namespace, e.g. ```cf ic namespace get```
    * CONTROLLER_HOSTNAME should be the (globally unique) cf route to be attached to the controller
    * ENABLE_SERVICEDISCOVERY determines whether to use the Bluemix-provided [Service Discovery](https://console.ng.bluemix.net/docs/services/ServiceDiscovery/index.html)
      instead of the A8 registry. When set to false, you can deploy your own customized A8 registry (not yet implemented).
    * ENABLE_MESSAGEHUB determines whether to use the Bluemix-provided [Message Hub](https://console.ng.bluemix.net/docs/services/MessageHub/index.html#messagehub).
      When set to false, the A8 proxies will use a slower polling algorithm to get changes from the A8 Controller.  
      Note that the Message Hub Bluemix service is not a free service, and using it might incur costs.
    * ...

1. Deploy the A8 controlplane by running [bluemix/deploy-controlplane.sh](bluemix/deploy-controlplane.sh).
    Verify that the controller is running by ```cf ic group list``` and checking if the ```amalgam8_controller``` group is running.

1. Deploy the Bookinfo app by running [bluemix/deploy-bookinfo.sh](bluemix/deploy-bookinfo.sh)

1. Configure the Amalgam8 CLI according to the routes defined in [.bluemixrc file](bluemix/.bluemixrc)

    ```
    export A8_CONTROLLER_URL=https://amalgam8-controller.mybluemix.net
    ```

1. Confirm the microservices are running

    ```
    a8ctl service-list
    ```
    
    Should produce the following output:
    
    ```
    +-------------+---------------------+
    | Service     | Instances           |
    +-------------+---------------------+
    | productpage | v1(1)               |
    | ratings     | v1(1)               |
    | details     | v1(1)               |
    | reviews     | v1(1), v2(1), v3(1) |
    +-------------+---------------------+
     ```

1. Route all traffic to version v1 of each microservice

    ```bash
    a8ctl route-set productpage --default v1
    a8ctl route-set ratings --default v1
    a8ctl route-set details --default v1
    a8ctl route-set reviews --default v1
    ```

1. Confirm the routes are set by running the following command

    ```bash
    a8ctl route-list
    ```

    You should see the following output:

    ```
    +-------------+-----------------+-------------------+
    | Service     | Default Version | Version Selectors |
    +-------------+-----------------+-------------------+
    | ratings     | v1              |                   |
    | productpage | v1              |                   |
    | details     | v1              |                   |
    | reviews     | v1              |                   |
    +-------------+-----------------+-------------------+
    ```

    Open the ${BOOKINFO_URL}/productpage/productpage from your browser and you should see the bookinfo application displayed.  
    (Replace BOOKINFO_URL with the value defined in [.bluemixrc file](bluemix/.bluemixrc))
  
1. Now that the application is up and running, you can try out the other a8ctl commands as described in
    [test & deploy demo](https://github.com/amalgam8/examples/blob/master/demo-script.md)
   

## Amalgam8 on Google Cloud Platform <a id="gcp"></a>

1. Setup [Google Cloud SDK](https://cloud.google.com/sdk/) on your machine

1. Setup a cluster of 3 nodes

1. Launch the control plane services

    ```bash
    kubernetes/run-controlplane-gcp.sh start
    ```

1. Locate the node where the controller is running and assign an
   external IP to the node if needed

1. Initialize the first tenant. The `run-controlplane-gcp.sh` script stores
   the JSON payload to initialize the tenant in the `TENANT_REG` environment variable.

    ```bash
    echo $TENANT_REG|curl -H "Content-Type: application/json" -d @- http://ControllerExternalIP:31200/v1/tenants'
    ```

1. Deploy the API gateway

    ```bash
    kubectl create -f kubernetes/gateway.yaml
    ```

    Obtain the public IP of the node where the gateway is running. This will be
    the be IP at which the sample app will be accessible.

1. You can now deploy the sample apps as described in "Running the sample
    apps" section above. Remember to replace the IP address `localhost`
    with the public IP address of the node where the gateway service is
    running on the Google Cloud Platform.

1. Visualizing your deployment with Weave Scope

    ```bash
    kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false
    ```

    Once weavescope is up and running, you can view the weavescope dashboard
    on your local host using the following commands
  
    ```bash
    kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath={.items..metadata.name}) 4040
    ```
  
    You can open http://localhost:4040 on your browser to access the Scope UI.
