# Getting Started with Amalgam8 using Docker Compose

## Overview

We will first deploy the Amalgam8 control plane and then a sample web
application made of 4 microservices. We will then use Amalgam8's control
plane to accomplish the following tasks:

1. Route traffic to specific versions of microservices using the `a8ctl`
   command line client that interacts with the A8 Control Plane.
2. Exercise the resiliency testing capabilities in Amalgam8 using the
   Gremlin framework to conduct systematic resilience testing, i.e., inject
   reproducible failure scenarios and run automated assertions on recovery
   behavior of the microservices. Specifically,
   * (Ad-hoc approach) Inject failures in the call path between two
   microservices while restricting the failure impact to a test user. You
   (the test user) would notice (manual) that the application is failing in
   an unexpected way.
   * (Systematic approach) Using the Gremlin framework (automated) to
     inject the same failure and verify whether the microservices recover
     in the expected manner.
3. Exercise the version routing capabilities in Amalgam8 by gradually
   increasing traffic from an old to a new version of an internal
   microservice.


## Pre-requisites

To run in a local docker environemnt, you can either use the vagrant sandbox
or you can simply install [Docker](https://docs.docker.com/engine/installation/),
[Docker Compose](https://docs.docker.com/compose/install/),
and the [Amalgam8 CLI](https://pypi.python.org/pypi/a8ctl)
on your own machine.

## Start the multi-tenant control plane and a tenant

Start the control plane services (registry and controller) by running the
following command:

```
compose/run-controlplane-docker.sh start
```

The above command also creates a tenant named "local" in the
control plane. 

Before we start using the `a8ctl` command line utility, we need to point it
to the address of the controller and the registry.
* If you are running the docker setup using the Vagrant file in the
`examples` folder, then set the following environment variables:

```bash
export A8_CONTROLLER_URL=http://192.168.33.33:31200
export A8_REGISTRY_URL=http://192.168.33.33:31300
```

* If you are running Docker locally,

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


You can confirm everything is working with the following command:

```bash
a8ctl service-list
```

The command shouldn't return any services, since we haven't started any yet, 
but if it returns the follwoing empty table, the control plane services (and CLI) are working as expected:

```
+---------+-----------------+-------------------+
| Service | Default Version | Version Selectors |
+---------+-----------------+-------------------+
+---------+-----------------+-------------------+
```


## Deploy the tenant application

Every tenant application should have
an [API Gateway](http://microservices.io/patterns/apigateway.html) that
provides a single user-facing entry point for a microservices-based
application.  You can control the Amalgam8 gateway for different purposes,
such as version routing, red/black deployments, canary testing, resiliency
testing, and so on. The Amalgam8 gateway is a simple lightweight Nginx
server that is controlled by the control plane.


### Deploy the API gateway

To start the API gateway, run the following command:

```bash
docker-compose -f compose/gateway.yaml up -d
```

Usually, the API gateway is mapped to a DNS route. However, in our local
standalone environment, you can access it by using the fixed IP address and
port (http://192.168.33.33:32000), which was pre-configured for the sandbox
environment. If you are using docker directly, then the gateway should be
accessible at http://localhost:32000 or http://dockermachineip:32000 .

Confirm that the API gateway is running by accessing the URL
from your browser. If all is well, you should
see a simple **Welcome to nginx!** page in your browser.

**Note:** You only need one gateway per tenant. A single gateway can front more
than one application under the tenant at the same time, so long as they
don't implement any conflicting microservices.

### Run the [Bookinfo sample app](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md)

An overview of the Bookinfo application can be found under the
[Bookinfo app](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md). Bring
up the bookinfo sample app by running the following command:

```bash
docker-compose -f bookinfo.yaml up -d
```

Confirm that the microservices are running, by running the following command:

```bash
a8ctl service-list
```

The following 4 microservices are displayed:

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

### Route all traffic to version v1 of each microservice:

Route all of the incoming traffic to version v1 only for each service, by
running the following command:


```bash
a8ctl route-set productpage --default v1
a8ctl route-set ratings --default v1
a8ctl route-set details --default v1
a8ctl route-set reviews --default v1
```

Confirm the routes are set by running the following command:

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

Open http://192.168.33.33:32000/productpage/productpage from your browser
and you should see the bookinfo application displayed. Notice that the
product page is displayed, with no rating stars since `reviews:v1` does not
access the ratings service.

**Note**: Remember to replace the IP address above with the appropriate IP
for your environment, i.e., http://localhost:32000 or
http://dockermachineip:32000 

### Version-aware routing

Lets enable the ratings service for test user "jason" by routing productpage
traffic to `reviews:v2` instances.

```bash
a8ctl route-set reviews --default v1 --selector 'v2(user="jason")'
```

Confirm the routes are set:

```
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
| reviews     | v1              | v2(user="jason")  |
+-------------+-----------------+-------------------+
```

Log in as user "jason" at the `productpage` web page.
You should now see ratings (1-5 stars) next to each review.

**Note**: Remember to replace the IP address above with the appropriate IP
for your environment, i.e., http://localhost:32000 or
http://dockermachineip:32000 

### Resilience Testing with Gremlin

[Gremlin](https://github.com/ResilienceTesting/gremlinsdk-python) is a
framework for systematically testing the ability of microservice-based
applications to recover from user-defined failure scenarios. In addition to
the fault injection, it allows the developer to script the set of
assertions that must be satisfied by the application as a whole when
recovering from the failure. In this demo we are only going to demonstrate
a couple of simple features to highlight the underlying support for Gremlin
provided by the Amalgam8 runtime.

* In the bookinfo application, the *reviews:v2 service has a 10 second timeout
for its calls to the ratings service*. To test that the end-to-end flow
works under normal circumstances, we are going to *inject a 7 second delay*
into the ratings microservice call from the reviews microservice, to make
sure that all microservices in the call chain function correctly.

* Lets add a fault injection rule via the `a8ctl` CLI that injects a 7s
delay in all requests with an HTTP Cookie header value for the user
"jason".

```bash
a8ctl rule-set --source reviews --destination ratings --header Cookie --pattern 'user=jason' --delay-probability 1.0 --delay 7
```

Verify the rule has been set by running this command:

```bash
a8ctl rule-list
```

You should see the following output:

```
+---------+-------------+--------+----------------+-------------------+-------+-------------------+------------+
| Source  | Destination | Header | Header Pattern | Delay Probability | Delay | Abort Probability | Abort Code |
+---------+-------------+--------+----------------+-------------------+-------+-------------------+------------+
| reviews | ratings     | Cookie | .*?user=jason  | 1                 | 7     | 0                 | 0          |
+---------+-------------+--------+----------------+-------------------+-------+-------------------+------------+
```

* Lets see the fault injection in action. Ideally the frontpage of the
  application should take 7+ seconds to load. To see the web page response
  time, open the *Developer Tools* (IE, Chrome or Firefox). The typical key
  combination is (Ctrl+Shift+I) for Windows and (Alt+Cmd+I) in Mac.

Reload the `productpage` web page.

You will see that the webpage loads in about 6 seconds. The reviews section
will show *Sorry, product reviews are currently unavailable for this book*.

Something is not working as expected in the application. If the reviews
service has a 10 second timeout, the product page should have returned
after 7 seconds with full content. What we see however is that the entire
reviews section is unavailable.

Notice that we are restricting the failure impact to user Jason only. If
you login as any other user, say "shriram" or "frank", you would not
experience any delays.

#### Use a Gremlin Recipe to systematically test the application

We'll now use a *gremlin recipe* that describes the application topology
(`topology.json`), reproduces the (7 seconds delay) failure scenario (`gremlins.json`),
and adds a set of assertions (`checklist.json`)
that we expect to pass: each service in the call chain should return `HTTP
200 OK` and the productpage should respond in 7 seconds.

* Edit the IP address of the `log_server` field in the file
  `apps/bookinfo/checklist.json` to point to the IP address where the
  controller is running (192.168.33.33 or localhost or your docker machine
  IP).

* Run the recipe using the following command from the main examples folder:

```bash
a8ctl recipe-run --topology apps/bookfino/topology.json --scenarios apps/bookinfo/gremlins.json --checks apps/bookinfo/checklist.json --header 'Cookie' --pattern='user=jason'
```

You should see the following output:

```
Inject test requests with HTTP header Cookie: value, where value matches the pattern user=jason
When done, press Enter key to continue to validation phase
```

Normally, we would have automated tools to inject load into the
application. In this case, for the purpose of this demo walkthrough, we
will manually inject load into the application. When logged in as user
jason, reload the `productpage` web page to once again run the
scenario, and then press Enter on the console where the above command was
run.

Expected output:

```
+-----------------------+-------------+-------------+--------+-----------------------+
| AssertionName         | Source      | Destination | Result | ErrorMsg              |
+-----------------------+-------------+-------------+--------+-----------------------+
| bounded_response_time | gateway     | productpage | PASS   |                       |
| http_status           | gateway     | productpage | PASS   |                       |
| http_status           | productpage | reviews     | FAIL   | unexpected status 499 |
| http_status           | reviews     | ratings     | PASS   |                       |
+-----------------------+-------------+-------------+--------+-----------------------+
Cleared fault injection rules from all microservices
```

**Understanding the output:** The above output indicates that the
productpage microservice timed out on its API call to the reviews
service. This indication is from status code HTTP 499, which is Nginx's
code to indicate that the caller closed its TCP connection
prematurely. However, we also see that the call from reviews to ratings
service was successful! This behavior suggests that the *productpage service
has a smaller timeout to the reviews service, compared to the timeout
duration between the reviews and ratings service.*

What we have here is a typical bug in microservice applications:
**conflicting failure handling policies in different
microservices**. Gremlin's systematic resilience testing approach enables
us to spot such issues in production deployments without impacting real
users.

**Fixing the bug** At this point we would normally fix the problem by
either increasing the productpage timeout or decreasing the reviews to
ratings service timeout, terminate and restart the fixed microservice, and
then run a gremlin recipe again to confirm that the productpage returns its
response without any errors.  (Left as an exercise for the reader - change
the gremlin recipe to use a 2.8 second delay and then run it against the v3
version of reviews.)

However, we already have this fix running in v3 of the reviews service, so
we can next demonstrate deployment of a new version.

### Gradually migrate traffic to reviews:v3 for all users

Now that we have tested the reviews service, fixed the bug and deployed a
new version (`reviews:v3`), lets route all user traffic from `reviews:v1`
to `reviews:v3` in a gradual manner.

First, stop any `reviews:v2` traffic:

```bash
a8ctl route-set reviews --default v1
```

Now, transfer traffic from `reviews:v1` to `reviews:v3` with the following series of commands:

```bash
a8ctl traffic-start reviews v3
```

You should see:
```
Transfer starting for reviews: diverting 10% of traffic from v1 to v3
```

Things seem to be going smoothly. Lets increase traffic to reviews:v3 by another 10%.

```bash
a8ctl traffic-step reviews
```

You should see:
```
Transfer step for reviews: diverting 20% of traffic from v1 to v3
```

Lets route 50% of traffic to `reviews:v3`

```bash
a8ctl traffic-step reviews --amount 50
```

We are confident that our Bookinfo app is stable. Lets route 100% of traffic to `reviews:v3`
```bash
a8ctl traffic-step reviews --amount 100
```

You should see:
```
Transfer complete for reviews: sending 100% of traffic to v3
```

If you log in to the `productpage` as any
user, you should see book reviews with *red* colored star ratings for each
review.

### Cleanup to restart demo

```bash
docker-compose -f compose/bookinfo.yaml kill
docker-compose -f compose/bookinfo.yaml rm -f

docker-compose -f compose/gateway.yaml kill
docker-compose -f compose/gateway.yaml rm -f

compose/run-controlplane-docker.sh stop
```

----

### TL;DR - Command summary (post control plane deployment)

```
docker-compose -f compose/gateway.yaml up -d
docker-compose -f compose/bookinfo.yaml up -d
a8ctl service-list

a8ctl route-set productpage --default v1
a8ctl route-set ratings --default v1
a8ctl route-set details --default v1
a8ctl route-set reviews --default v1
a8ctl route-list

a8ctl route-set reviews --default v1 --selector 'v2(user="jason")'
a8ctl route-list

a8ctl rule-set --source reviews --destination ratings --header Cookie --pattern 'user=jason' --delay-probability 1.0 --delay 7
a8ctl rule-list
a8ctl recipe-run --topology topology.json --scenarios gremlins.json --checks checklist.json --header 'Cookie' --pattern='user=jason'

a8ctl route-set reviews --default v1
a8ctl traffic-start reviews v3 # 10%
a8ctl traffic-step reviews # 20%
a8ctl traffic-step reviews --amount 50 # 50%
a8ctl traffic-step reviews --amount 100 # 100%
```
