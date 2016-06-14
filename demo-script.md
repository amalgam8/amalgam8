# Amalgam8 Test and Deploy Demo

## Overview

The test and deploy demo is a walkthrough end-to-end example, to test out some features so that you can see how you might want to use Amalgam8.

Demo goals:

1. Deploy a simple web application consisting of 4 microservices using
   kubernetes.
2. Route traffic to specific versions of microservices using the `a8ctl`
   command line client that interacts with the A8 Control Plane.
3. Exercise the resiliency testing capabilities in Amalgam8 using the
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
5. Exercise the version routing capabilities in Amalgam8 by gradually
   increasing traffic from an old to a new version of an internal
   microservice.


## Bring up a Kubernetes cluster in Vagrant

To get started, install a recent version of [Vagrant](https://www.vagrantup.com/downloads.html) and follow the steps below.

1. Clone the Amalgam8 example repo and start the vagrant environment.

```bash
git clone https://github.com/amalgam8/examples

cd examples
vagrant up
vagrant ssh

cd $GOPATH/src/github.com/amalgam8
```

**Note:** If you stopped a previous Vagrant VM and restarted it, Kubernetes
might not run correctly. If you have problems, try uninstalling Kubernetes
by running the following command:
  
```bash
sudo examples/uninstall-kubernetes.sh
```

Then re-install Kubernetes, by running the following command:

```bash
sudo examples/install-kubernetes.sh
```

## Start the multi-tenant control plane and a tenant

Start the control plane services (registry and controller) by running the
following command:

```bash
examples/run-controlplane-local.sh start
```

The above command also creates a tenant named "local" in the
control plane. 

You can confirm everything is up and running with the following command:

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

You can also access the registry at http://192.168.33.33:5080 from the host machine
(outside the vagrant box), and the controller at http://192.168.33.33:31200.
To access the control plane details of tenant *local*, access
http://192.168.33.33:31200/v1/tenants/local/ from your browser.

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
kubectl create -f examples/gateway/gateway.yaml
```

Usually, the API gateway is mapped to a DNS route. However, in our local
standalone environment, you can access it by using the fixed IP address and
port (http://192.168.33.33:32000), which was pre-configured for the sandbox
environment.

Confirm that the API gateway is running by accessing the
http://192.168.33.33:32000 from your browser. If all is well, you should
see a simple **Welcome to nginx!** page in your browser.

**Note:** You only need one gateway per tenant. A single gateway can front more
than one application under the tenant at the same time, so long as they
don't implement any conflicting microservices.

Confirm that the control plane and API gateway are active by running the
following command:

```bash
kubectl get po
```

The returned list should include at least the following 5 pods:

```
NAME                   READY     STATUS    RESTARTS   AGE
controller-yab4n       1/1       Running   0          55m
gateway-dzh1w          1/1       Running   0          55m
kafka-s7xvb            1/1       Running   0          55m
logserver-gkpbw        3/3       Running   0          55m
registry-aat8k         1/1       Running   0          55m
```

### Run the [Bookinfo sample app](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md)

An overview of the Bookinfo application can be found under the
[Bookinfo app](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md). Bring
up the bookinfo sample app by running the following command:

```bash
kubectl create -f examples/apps/bookinfo/bookinfo.yaml
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

Log in as user "jason" at the url http://192.168.33.33:32000/productpage/productpage
You should now see ratings (1-5 stars) next to each review.

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

Reload the web page (http://192.168.33.33:32000/productpage/productpage).

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

* Run the recipe using the following command:

```bash
a8ctl recipe-run --topology topology.json --scenarios gremlins.json --checks checklist.json --header 'Cookie' --pattern='user=jason'
```

You should see the following output:

```
Inject test requests with HTTP header Cookie: value, where value matches the pattern user=jason
When done, press Enter key to continue to validation phase
```

Normally, we would have automated tools to inject load into the
application. In this case, for the purpose of this demo walkthrough, we
will manually inject load into the application. When logged in as user
jason, reload the webpage
(http://192.168.33.33:32000/productpage/productpage) to once again run the
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

If you log in to http://192.168.33.33:32000/productpage/productpage as any
user, you should see book reviews with *red* colored star ratings for each
review.

----

### TL;DR - Command summary (post K8S and control plane deployment)

```
kubectl create -f examples/apps/bookinfo/bookinfo.yaml
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

#### Cleanup to restart demo

```
kubectl delete -f examples/apps/bookinfo.yaml
a8ctl route-delete productpage
a8ctl route-delete ratings
a8ctl route-delete details
a8ctl route-delete reviews
```

To (re)start the control plane:

If `kubectl get svc` not working, then

```bash
sudo examples/uninstall-kubernetes.sh
sudo examples/install-kubernetes.sh
```

To stop an operational control plane and the API gateway

```bash
examples/controlplane/run-controlplane-local.sh stop
kubectl delete -f examples/gateway/gateway.yaml
```

To start control plane and gateway

```bash
examples/controlplane/run-controlplane-local.sh start
kubectl create -f examples/gateway/gateway.yaml
```
