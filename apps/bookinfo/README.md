# Amalgam8 Bookinfo sample

## Overview

The bookinfo sample is a simple application that implements a web page that displays information about a book, 
similar to a single catalog entry of an online book store. Displayed on the page is a description of the book,
book details (ISBN, number of pages, and so on), and a few book reviews.

The sample application is broken into four separate microservices:

* *productpage*. The productpage microservice calls the *details* and *reviews* microservices to populate the page. It provides a good example to experiment with both mid-tier and edge service routing.
* *details*. The details microservice contains book information.
* *reviews*. The reviews microservice contains book reviews. It also calls the *ratings* microservice, to provide two levels of downstream mid-tier routing.
* *ratings*. The ratings microservice contains book ranking information that accompanies a book review. 

There are 3 versions of the reviews microservice:

* Version v1 doesn't call the ratings service.
* Version v2 calls the ratings service, and displays each rating as 1 to 5 black stars.
* Version v3 calls the ratings service, and displays each rating as 1 to 5 red stars.

![Microservice dependencies](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/dependencies.jpg)

## Running the bookinfo demo

In this demo, we will use Amalgam8's control plane to accomplish the following tasks:

1. Route traffic to specific versions of a microservice by using the `a8ctl`
   CLI that interacts with the Amalgam8 control plane.

2. Demonstrate the resiliency testing capability of Amalgam8 by using the Gremlin framework to introduce failure scenarios. These scenarios can be reproduced, and automated test conditions applied in the following methods:

   * *Ad-hoc*: Inject failures manually to the call path between two
   microservices while restricting the failure impact to a user. The user would notice in using the application that the application is failing in an unexpected way.
   * *Systematic*: Use the Gremlin framework to automate injecting the same failure, and verify whether the microservices recover as expected.

3. Demonstrate the version routing capabilities in Amalgam8 by gradually
   increasing traffic from an old to a new version of an internal
   microservice.

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

Confirm that the microservices are running, using the following command:

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

Open http://localhost:32000/productpage/productpage from your browser
and you should see the bookinfo application `productpage` displayed.
Notice that the `productpage` is displayed, with no rating stars since `reviews:v1` does not
access the ratings service.

**Note**: Replace localhost:32000 above with the appropriate host
for your environment (for example, "bookinfo.mybluemix.net", etc.).

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
a8ctl recipe-run --topology apps/bookinfo/topology.json --scenarios apps/bookinfo/gremlins.json --checks apps/bookinfo/checklist.json --header 'Cookie' --pattern='user=jason'
```

You should see the following output:

```
Inject test requests with HTTP header Cookie: value, where value matches the pattern user=jason
When done, press Enter key to continue to validation phase
```

Normally, we would have automated tools to inject load into the
application. In this case, for the purpose of this demo walkthrough, we
will manually inject load into the application. When logged in as user
`jason`, reload the `productpage` web page to once again. Wait a few seconds
seconds to let the logs propagate from the app/sidecar containers to the
logstash server and finally to elasticsearch. Then, press Enter on the console
where the above command was run.

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

*Note:* When logs from logstash do not appear in elasticsearch by the
time you hit the Enter key, one or more tests above might fail with the
error message `No log entries found`. When you encounter this situation,
re-run the above command (`a8ctl recipe-run ...`) and wait for a longer
time before hitting the Enter key. We are working on a cleaner fix to
address the log propagation delay problem.

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
