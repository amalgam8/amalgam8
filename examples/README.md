# Amalgam8 Examples

Sample microservice-based applications and local sandbox environment for Amalgam8.

An overview of Amalgam8 can be found at www.amalgam8.io.

[//]: # (**Note:** This is an unstable branch. If you are experimenting with Amalgam8 for the first time, please use the stable branch (`git checkout -b 0.1.0 origin/0.1.0`) and use this [README](https://github.com/amalgam8/amalgam8/examples/blob/0.2.0/README.md) from the stable branch.)

## Overview <a id="overview"></a>

This project includes a number of Amalgam8 sample programs, scripts and a preconfigured environment to allow
you to easily run, build, and experiment with the provided samples, in several environments.
In addition, the scripts are generic enough that you can easily deploy
the samples to other environments as well.

The following samples are available for Amalgam8:

* [Helloworld](apps/helloworld/) is a single microservice app that demonstrates how to route traffic to different versions of the same microservice
* [Bookinfo](apps/bookinfo/) is a multiple microservice app used to demonstrate and experiment with several Amalgam8 features

### Setup

Before running the samples, you need to setup the requisite environment.

* *Vagrant sandbox*: The repository's root directory includes a Vagrant file that provides an environment with everything needed to run, and build, the samples ([Go](http://golang.org/), [Docker](http://www.docker.com/), [Kubernetes](http://kubernetes.io/), [Amalgam8 CLI w/ Gremlin SDK](https://github.com/amalgam8/amalgam8/a8ctl) already installed. Depending on the runtime environement you want to try, using it may be the easiest way to get Amalgam8 up and running.

* *Custom setup*: If you are not using the vagrant environment, then install the following pre-requisites:
  * Amalgam8 python CLI
   ```bash
   sudo pip install git+https://github.com/amalgam8/amalgam8/a8ctl
   ```
  * [Docker 1.10 or later](https://docs.docker.com/engine/installation/)
  * [Docker Compose 1.5.1 or later](https://docs.docker.com/compose/install/)

* *Development Mode*: If you'd like to also be able to change and compile the code, or build the images,
refer to the [Developer Instructions](../devel/).

### Deployment Options

Amalgam8 platform can be deployed on the following container runtimes and PaaS environments. Pick an option below and follow the instructions in the respective section.

* Localhost Deployment
    * [Docker](#local-docker)
    * [Kubernetes](#local-k8s)
    * [Marathon/Mesos](#local-marathon)
* Cloud Deployment
    * [IBM Bluemix](#bluemix)
    * [Google Compute Cloud](#gcp)

While Amalgam8 supports multi-tenancy, for the sake of simplicity, this walk through will setup Amalgam8 in single-tenant mode.

## Amalgam8 with Docker - local environment <a id="local-docker"></a>

The installation steps below have been tested with the Vagrant sandbox
environment (based on Ubuntu 14.04) as well as with Docker for Mac Beta
(v1.11.2-beta15 or later). These steps have not been tested on Docker for
Windows Beta.

The following instructions assume that you are using the Vagrant
environment. Where appropriate, environment specific instructions are
provided.

1. Download the [Vagrantfile](Vagrantfile) and start the Vagrant environment (or install and setup the equivalent dependencies manually).

    ```bash
    vagrant up
    vagrant ssh
    ```

    Once inside the vagrant box, switch to the Amalgam8 examples folder:

    ```bash
    cd amalgam8/examples
    ```

1. Start the control plane services (service registry, controller) and the ELK stack

    ```bash
    docker/run-controlplane-docker.sh start
    ```


1. The `a8ctl` command line utility can be used to setup routes between microservices. 
   Before we start using the `a8ctl` command line utility, we need to set 
   the `A8_CONTROLLER_URL` and the `A8_REGISTRY_URL` environment variables
   to point to the addresses of the controller and the registry respectively.

    * If you are running the Docker setup using the Vagrant file in the
    `examples` folder or if you are running Docker locally (on Linux or
    Docker for Mac Beta)

    ```bash
    export A8_CONTROLLER_URL=http://localhost:31200
    export A8_REGISTRY_URL=http://localhost:31300
    ```

    * If you are running Docker using the Docker Toolbox with Docker Machine,
    then set the environment variables to the IP address of the VM created
    by Docker Machine. For example, assuming you have only one Docker
    Machine running on your system, the following commands will setup the
    appropriate environment variables:

    ```bash
    export A8_CONTROLLER_URL=`docker-machine ip`
    export A8_REGISTRY_URL=`docker-machine ip`
    ```

1. Confirm everything is working with the following command:

    ```bash
    a8ctl service-list
    ```

    The command shouldn't return any services, since we haven't started any yet.
    It should return the following empty table confirming that the control plane services (and CLI) are working as expected:

    ```
    +---------+-----------+
    | Service | Instances |
    +---------+-----------+
    +---------+-----------+
    ```


1. Deploy the API gateway

    The [API Gateway](http://microservices.io/patterns/apigateway.html) 
    provides a single user-facing entry point for a microservices-based
    application.  You can use the Amalgam8 gateway for different purposes,
    such as version-aware routing, red/black deployments, canary testing, resiliency
    testing, and so on. The Amalgam8 API gateway is a simple lightweight Openresty server (i.e. Nginx)
    that is controlled by the control plane.


    To start the API gateway, run the following command:

    ```bash
    docker-compose -f docker/gateway.yaml up -d
    ```

    Usually, the API gateway is mapped to a DNS route. However, in our local
    standalone environment, you can access it at port 32000 on localhost.
    If you are using Docker directly, then the gateway should be
    accessible at `http://localhost:32000` or `http://dockermachineip:32000`.

1. Confirm that the API gateway is running by accessing
    `http://localhost:32000` from your browser.
    If all is well, you should see a simple **Welcome to nginx!**
    page in your browser.

1. Follow the instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:

    ```bash
    docker-compose -f docker/helloworld.yaml up -d
    docker-compose -f docker/helloworld.yaml scale helloworld-v1=2
    docker-compose -f docker/helloworld.yaml scale helloworld-v2=2
    ```
        
    * Follow the instructions for the [Helloworld](apps/helloworld/) example

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

    * Follow the instructions for the [Bookinfo](apps/bookinfo/) example

    * To shutdown the bookinfo instances, run the following commands:
    
    ```
    docker-compose -f docker/bookinfo.yaml kill
    docker-compose -f docker/bookinfo.yaml rm -f
    ```

    When you are finished, shut down the gateway and control plane services by running the following commands:

    ```
    docker/cleanup.sh
    ```

## Amalgam8 with Kubernetes - local environment <a id="local-k8s"></a>

The following setup has been tested with Kubernetes v1.2.3.

1. Download the [Vagrantfile](Vagrantfile) and start the Vagrant environment (or install and setup the equivalent dependencies manually).

    ```bash
    vagrant up
    vagrant ssh
    ```

    Once inside the vagrant box, switch to the Amalgam8 examples folder and setup the required environment variables

    ```bash
    cd amalgam8/examples
    export A8_CONTROLLER_URL=http://localhost:31200
    export A8_REGISTRY_URL=http://localhost:31300
    ```

    Start Kubernetes, by running the following command:
 
    ```bash
    sudo kubernetes/install-kubernetes.sh
    ```

    **Note:** If you stopped a previous Vagrant VM and restarted it, Kubernetes might be started already, but in a bad state.
    If you have problems, first start by removing previously deployed
    services and uninstalling Kubernetes with the following commands: 
 
    ```bash
    kubernetes/cleanup.sh
    sudo kubernetes/uninstall-kubernetes.sh
    ```

    Alternatively, you can install a local version of Kubernetes via other options such as [Minikube](https://github.com/kubernetes/minikube). In this case, make sure to clone the [amalgam8 repository](https://github.com/amalgam8/amalgam8) and install the Python-based `a8ctl` CLI utility on your host machine.

1. Start the local control plane services (registry and controller) and the ELK stack by running the following script:

    ```bash
    kubernetes/run-controlplane-local-k8s.sh start
    ```

1. Confirm that the control plane is working:

    ```bash
    a8ctl service-list
    ```

    The command shouldn't return any services, since we haven't started any yet.
    It should return the following empty table confirming that the control plane services (and CLI) are working as expected:
    
    ```
    +---------+-----------+
    | Service | Instances |
    +---------+-----------+
    +---------+-----------+
    ```
    
    **Note:** If this did not work, it's probabaly because the image download and/or service initialization took too long.
    This is usually fixed by waiting a minute or two, and then running `kubernetes/run-controlplane-local-k8s.sh stop` and then
    repeating the previous step.
    
    You can also access the registry at `http://localhost:31300` from the host machine
    (outside the Vagrant box), and the controller at `http://localhost:31200` .
    To access the control plane details of tenant *local*, access
    http://localhost:31200/v1/tenants/local from your browser.


1. Run the [API Gateway](http://microservices.io/patterns/apigateway.html) with the following commands:

    ```bash
    kubectl create -f kubernetes/gateway.yaml
    ```
    
    Usually, the API gateway is mapped to a DNS route. However, in our local
    standalone environment, you can access it at port 32000 on localhost.

1. Confirm that the API gateway is running by accessing
    `http://localhost:32000` from your browser. If all is well, you should
    see a simple **Welcome to nginx!** page in your browser.

1. Visualize your deployment using Weave Scope by accessing
   `http://localhost:30040` . Click on `Pods` tab. You should see a graph of
   pods depicting the connectivity between them. As you create more apps
   and manipulate routing across microservices, the graph changes in
   real-time.
   
1. <a id="local-k8s-samples">Deploying sample apps:</a> follow the instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:
    
        ```bash
        kubectl create -f kubernetes/helloworld.yaml
        ```
        
    * Follow the instructions for the [Helloworld](apps/helloworld/) example
 
    * To shutdown the helloworld instances, run the following command:
    
        ```bash
        kubectl delete -f kubernetes/helloworld.yaml
        ```

    (b) **bookinfo** sample

    * Start the bookinfo application:
    
        ```bash
        kubectl create -f kubernetes/bookinfo.yaml
        ```

    * Follow the instructions for the [Bookinfo](apps/bookinfo/) example
    
    * To shutdown the bookinfo instances, run the following command:
    
        ```bash
        kubectl delete -f kubernetes/bookinfo.yaml
        ```

1. When you are finished, shut down the gateway and control plane services by running the following commands:

    ```bash
    kubernetes/cleanup.sh
    ```

## Amalgam8 with Marathon/Mesos - local environment <a id="local-marathon"></a>

The following setup has been tested with Marathon 0.15.2 and Mesos 0.26.0.

1. Download the [Vagrantfile](Vagrantfile)
2. **Edit the Vagrant file** and add the following line inside the `config.vm` block:

    ```ruby
    config.vm.network "private_network", ip: "192.168.33.33/24"
    ```

3. Start the Vagrant environment.

    ```bash
    vagrant up
    vagrant ssh
    ```

    Once inside the vagrant box, switch to the Amalgam8 examples folder and setup the required environment variables

    ```bash
    cd amalgam8/examples
    export A8_CONTROLLER_URL=http://192.168.33.33:31200
    export A8_REGISTRY_URL=http://192.168.33.33:31300
    ```

1. Start the local control plane services (registry and controller) and the ELK stack by running the following script:
   
    ```bash
    marathon/run-controlplane-marathon.sh start
    ```

    From your browser, confirm that the Marathon dashboard is accessible at
    `http://192.168.33.33:8080` and the Mesos dashboard at `http://192.168.33.33:5050`

    Verify that the controller and registry are running via the Marathon dashboard.

    The single host marathon/mesos deployment is based on Holiday Check's [mesos-in-the-box](https://github.com/holidaycheck/mesos-in-the-box).

1. Launch the API Gateway
    
    ```bash
    marathon/run-component.sh gateway start
    ```

1. Confirm that the API gateway is running by accessing the
    `http://192.168.33.33:32000` from your browser. If all is well, you should
    see a simple **Welcome to nginx!** page in your browser.

1. Follow the instructions for the sample that you want to run.

    (a) **helloworld** sample

    * Start the helloworld application:

    ```bash
    marathon/run-component.sh helloworld start
    ```
        
    * Follow the instructions for the [Helloworld](apps/helloworld/) example

    * To shutdown the helloworld instances, run the following commands:
   
    ```bash
    marathon/run-component.sh helloworld stop
    ```

    (b) **bookinfo** sample

    * Start the bookinfo application:
    
    ```bash
    marathon/run-component.sh bookinfo start
    ```

    * Follow the instructions for the [Bookinfo](apps/bookinfo/) example

    * To shutdown the bookinfo instances, run the following commands:
    
    ```bash
    marathon/run-component.sh bookinfo stop
    ```

    When you are finished, shut down the gateway and control plane services by running the following commands:

    ```bash
    marathon/cleanup.sh
    ```

## Amalgam8 on IBM Bluemix <a id="bluemix"></a>

To run the [Bookinfo sample app](apps/bookinfo/)
on IBM Bluemix, follow the instructions below. If you are not a Bluemix user, you can register at [bluemix.net](http://bluemix.net/).

1. Download the [Vagrantfile](Vagrantfile) and start the Vagrant environment (or install and setup the equivalent dependencies manually).

    ```bash
    vagrant up
    vagrant ssh
    ```

    Once inside the vagrant box, install the following additional dependencies:
    [CF CLI 6.12.0 or later](https://github.com/cloudfoundry/cli/releases),
    [Bluemix CLI 0.3.3 or later](https://clis.ng.bluemix.net/),

1. Switch to the Amalgam8 examples folder

    ```bash
    cd amalgam8/examples
    ```

1. Login to Bluemix and initialize the container environment using ```bluemix login``` and ```bluemix ic init```

1. Create Bluemix routes (DNS names) for the registry, controller and the bookinfo app's gateway:  
    ```cf create-route myspace mybluemix.net -n mya8-registry```
    ```cf create-route myspace mybluemix.net -n mya8-controller```
    ```cf create-route myspace mybluemix.net -n mya8-bookinfo```
    

1. Customize the [bluemixrc](bluemix/.bluemixrc) file in the following manner:
    * BLUEMIX_REGISTRY_NAMESPACE should be your Bluemix registry namespace, e.g. ```bluemix ic namespace-get```
    * BLUEMIX_REGISTRY_HOST should be the Bluemix registry hostname. This needs to be set only if you're targeting a Bluemix region other than US-South.
    * REGISTRY_HOSTNAME should be the route name assigned to the registry in the previous step
    * CONTROLLER_HOSTNAME should be the route name assigned to the controller in the previous step
    * BOOKINFO_HOSTNAME should be the route name assigned to the bookinfo gateway in the previous step
    * ROUTES_DOMAIN should be the domain used for the Bluemix routes (e.g., mybluemix.net)

1. Deploy the control plane services (registry and controller) on bluemix.

    ```bash
    bluemix/deploy-controlplane.sh
    ```

    Verify that the controller and registry are running using the following commands: 

    ```bash
    bluemix ic groups
    ```
 
    You should see the groups `amalgam8_controller` and `amalgam8_registry` listed in the output.

1. Configure the Amalgam8 CLI according to the routes defined in `.bluemixrc`. For example

    ```
    export A8_CONTROLLER_URL=http://mya8-controller.mybluemix.net
    export A8_REGISTRY_URL=http://mya8-registry.mybluemix.net
    ```

1. Run the following command to confirm the control plane is working:

    ```bash
    a8ctl service-list
    ```

    The command shouldn't return any services, since we haven't started any yet.
    It should return the following empty table confirming that the control plane services (and CLI) are working as expected:
    
    ```
    +---------+-----------+
    | Service | Instances |
    +---------+-----------+
    +---------+-----------+
    ```

1. Deploy the API Gateway and the Bookinfo app.

    ```bash
    bluemix/deploy-bookinfo.sh
    ```

    Follow the instructions for the [Bookinfo](apps/bookinfo/) example

    **Note 1:** When you reach the part where the tutorial instructs you to open, in your browser, the bookinfo application at
    `http://localhost:32000/productpage/productpage`, make sure to change `http://localhost:32000` to `http://${BOOKINFO_HOSTNAME}.mybluemix.net` (substitute BOOKINFO_HOSTNAME with the value defined in the `.bluemixrc` file).

    **Note 2:** The Bluemix version of the bookinfo sample app does not yet support running the Gremlin recipe.
    We are working on integrating the app with the Bluemix Logmet services (ELK stack), to enable support for running Gremlin recipes.

1. When you are finished, shut down the gateway and control plane servers by running the following commands:

    ```bash
    bluemix/kill-bookinfo.sh
    bluemix/kill-controlplane.sh
    ```

## Amalgam8 on Google Cloud Platform <a id="gcp"></a>

1. Setup [Google Cloud SDK](https://cloud.google.com/sdk/) on your machine

1. Setup a cluster of 3 nodes

1. Launch the control plane services

    ```bash
    kubernetes/run-controlplane-gcp.sh start
    ```

1. Locate the node where the controller and registry are running and assign an external IP to the node.

1. Deploy the API gateway

    ```bash
    kubectl create -f kubernetes/gateway.yaml
    ```

    Obtain the public IP of the node where the gateway is running. This will be
    the be IP at which the sample app will be accessible.

1. Visualizing your deployment with Weave Scope

    ```bash
    kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false
    ```

    Once weavescope is up and running, you can view the weavescope dashboard
    on your local host using the following commands
  
    ```bash
    kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath={.items..metadata.name}) 4040
    ```
  
    Open `http://localhost:4040` on your browser to access the Scope UI. Click on `Pods` tab. 
    You should see a graph of pods depicting the connectivity between them. As you create 
    more apps and manipulate routing across microservices, the graph changes in real-time.

1. You can now deploy the sample apps as described in [Deploying sample apps](#local-k8s-samples) 
    section under the [local Kubernetes installation](#local-k8s) instructions. Remember to replace the IP address `localhost`
    with the public IP address of the node where the gateway service is running on the Google Cloud Platform.

## Contributing

Contributions and feedback are welcome!
Proposals and pull requests will be considered. 
Please see the [CONTRIBUTING.md](../CONTRIBUTING.md) file for more information.
