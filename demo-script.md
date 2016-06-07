# Amalgam8 Test and Deploy Demo

## Overview

The test and deploy demo is a walkthrough end-to-end example, to test out some features so that you can see how you might want to use Amalgam8.

What you will see in the demo:

1. How to start and deploy (i.e., enable web traffic for) an application consisting of 4 microservices, using kubernetes and amalgam8 CLI commands.
2. How with amalgam8 we can send traffic for a specific user to a second version of one of the microservices without impacting other users in a production deployment.
3. How amalgam8 allows us to test the live application by injecting a delay in the call path between 2 of the microservices.
4. How amalgam8 enables using the Gremlin SDK to perform systematic resilience testing with reproducible failure scenarios and assertions to uncover the cause of the problem.
5. How the CLI can be used to deploy a new (fixed) version of a microservice by gradually transferring traffic using amalgam8 percentage-based routing.

### Check the controlplane and API gateway are running

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

1. Confirm that the controlplane and API gateway are active by running the following command:

  ```
    kubectl get po
  ```

  The returned list should include at least the following 5 pods:

  ```
    NAME                   READY     STATUS    RESTARTS   AGE
    controller-yab4n       1/1       Running   0          55m
    gateway-dzh1w          1/1       Running   0          55m
    kafka-s7xvb            1/1       Running   0          55m
    logserver-gkpbw        2/2       Running   0          55m
    registry-aat8k         1/1       Running   0          55m
  ```

### Run the [Bookinfo sample](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/README.md) application microservices

2. Install the bookinfo sample app by running the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/apps/bookinfo
    ./build-services.sh
    kubectl create -f ./bookinfo.yaml
  ```
  
  If you previously installed the bookinfo sample, remove the following line from the command: 

  ```
    ./build-services.sh
  ```

3. Confirm that the microservices are running, by running the following command:

  ```
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

### Route all traffic to version v1 of each service:

4. Route all of the incoming traffic to version v1 only for each service, by running the following command:


  ```
    a8ctl route-set productpage --default v1
    a8ctl route-set ratings --default v1
    a8ctl route-set details --default v1
    a8ctl route-set reviews --default v1
  ```

5. Confirm the routes are set by running the following command:

  ```
    a8ctl route-list
  ```

  The following output is expected:

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

6. Point your browser to http://192.168.33.33:32000/productpage/productpage

  The product page is displayed, with no ratings (stars).

### Route the reviews microservice traffic

  In this example, we will route the traffic for the reviews microservice for the user "jason" to v2 instances only.

7. Change the routing rules for the review microservice by running the following command:

  ```
    a8ctl route-set reviews --default v1 --selector 'v2(user="jason")'
  ```

  Confirm the routes are set:

  ```
    a8ctl route-list
  ```

  The following output is expected:

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

8. Log in as user "jason" at the url http://192.168.33.33:32000/productpage/productpage

  You should now see ratings (1-5 stars) next to each review.

### Systematically test the resiliency of the system with Gremlin

  [Gremlin](https://github.com/ResilienceTesting/gremlinsdk-python)
  is a powerful tool for end-to-end resiliency testing of microservice-based systems.
  However, in this demo we are only going to demonstrate a couple of simple features to highlight
  the underlying support provided by the Amalgam8 runtime.

  In the bookinfo application, the reviews service has a 10 second timeout for its calls to the
  ratings service. To test that the end-to-end flow works under normal circumstances,
  we are going to inject a 7 second delay into the ratings microservice
  call from the reviews microservice, to make sure that all microservices in the call chain function correctly.

9. Inject a 7 second delay in all requests with an HTTP Cookie header value for the user Jason, by running the following command:

  ```
    a8ctl rule-set --source reviews --destination ratings --header Cookie --pattern 'user=jason' --delay-probability 1.0 --delay 7
  ```

  You can verify the rule has been set by running this command:

  ```
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

10. If you are using Chrome, open Chrome inspector (Ctrl+Shift+I), or in other browsers, right-click in the web page and
select `inspect`.

11. Reload the web page (http://192.168.33.33:32000/productpage/productpage).

  You will see that the webpage loads in about 6 seconds. The reviews section will
  show "Sorry, product reviews are currently unavailable for this book".

  This shows that something is wrong in the application. If the
  reviews service has a 10 second timeout, the product page should have
  returned after 7 seconds with full content.

#### Use a Gremlin Recipe to stage failures and run assertions

  We'll now use a *gremlin recipe* that describes the application topology,
  the test scenario (i.e., 7 seconds delay), and a set of checks (assertions) that
  we expect to pass (i.e., that all the services in the call chain should
  return `HTTP 200 OK` and the productpage should respond in 7 seconds).

12. Run the following command:

  ```
    a8ctl recipe-run --topology topology.json --scenarios gremlins.json --checks checklist.json --header 'Cookie' --pattern='user=jason'
  ```

  Expected output:

  ```
    Inject test requests with HTTP header Cookie: value, where value matches the pattern user=jason
    When done, press Enter key to continue to validation phase
  ```

  Reload the webpage (http://192.168.33.33:32000/productpage/productpage) to
  once again run the scenario, and then press Enter on the console where the
  above command was run.

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

  The above output indicates that the productpage microservice timed out on its API call
  to the reviews service. This indication is from status code HTTP 499, which is Nginx's
  code to indicate that the caller closed its TCP connection
  prematurely. However, we also see that the call from reviews to ratings
  service was successful! This behavior suggests that the productpage service
  has a smaller timeout to the reviews service, compared to the timeout
  duration between the reviews and ratings service.

  The above scenario highlights one of the common issues in microservice
  applications: conflicting failure handling policies. The systematic
  resilience testing approach by Gremlin enables us to spot such issues in
  production deployments without impacting real users.

  At this point we would normally fix the problem by either increasing
  the productpage timeout or decreasing the reviews to ratings service timeout,
  terminate and restart the fixed microservice, and then run a gremlin recipe again
  to confirm that the productpage returns its response without any errors.
  (Left as an exercise for the reader - change the gremlin recipe to use
  a 2.8 second delay and then run it against the v3 version of reviews.)

  However, we already have this fix running in v3 of the reviews service, so we can next demonstrate active deploy.

### Stop traffic to v2 and then do active deploy of v3 to all users

  We will now rollout the new v3 version of the reviews service to all users who were previously on v2.

13. Stop v2 traffic by running the following commands:

  ```
    a8ctl route-set reviews --default v1
    Set routing rules for microservice reviews
  ```

14. Rollout v3, in a series of commands and reviewing the progress, in the following format:

  ```
    $ a8ctl rollout-start reviews v3
    Rollout starting for reviews: diverting 10% of traffic from v1 to v3

    $ a8ctl rollout-step reviews
    Rollout step for reviews: diverting 20% of traffic from v1 to v3

    $ a8ctl rollout-step reviews --amount 30
    Rollout step for reviews: diverting 50% of traffic from v1 to v3

    $ a8ctl rollout-step reviews --amount 50
    Rollout complete for reviews: sending 100% of traffic to v3

  ```

15. Log in to http://192.168.33.33:32000/productpage/productpage as any user

16. Confirm that you can see the v3 ratings (1-5 red stars)

END OF DEMO

## Command summary

```
cd $GOPATH/src/github.com/amalgam8/examples/apps/bookinfo
kubectl create -f bookinfo.yaml
a8ctl service-list

a8ctl route-set productpage --default v1
a8ctl route-set ratings --default v1
a8ctl route-set details --default v1
a8ctl route-set reviews --default v1
a8ctl route-list

a8ctl route-set reviews --default v1 --selector 'v2(user="jason")'
a8ctl route-list

... TODO: Gremlin part

a8ctl route-set reviews --default v1

a8ctl rollout-start reviews v3 # 10%
a8ctl rollout-step reviews # 20%
a8ctl rollout-step reviews --amount 30 # 50%
a8ctl rollout-step reviews --amount 50 # 100%
```

## Cleanup to restart demo

```
cd $GOPATH/src/github.com/amalgam8/examples/apps/bookinfo
kubectl delete -f bookinfo.yaml
a8ctl route-delete productpage
a8ctl route-delete ratings
a8ctl route-delete details
a8ctl route-delete reviews
```

To (re)start the control plane:

```
cd $GOPATH/src/github.com/amalgam8/examples

# if "kubectl get svc" not working
./uninstall-kubernetes
./install-kubernetes

# if previous run needs to be stopped
./controlplane/run-controlplane.sh stop
kubectl delete -f gateway/gateway.yaml

# Start control plane and gateway
./controlplane/run-controlplane.sh compile # if code changes need to be picked up
./controlplane/run-controlplane.sh start
kubectl create -f gateway/gateway.yaml
