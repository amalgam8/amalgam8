# Amalgam8

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/amalgam8
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/amalgam8
[Travis]: https://travis-ci.org/amalgam8/amalgam8
[Travis Widget]: https://travis-ci.org/amalgam8/amalgam8.svg?branch=master

## TL;DR

1. Watch this YouTube video:

<a href="http://www.youtube.com/watch?feature=player_embedded&v=gvjhrxwX7S8" target="_blank"><img
src="http://img.youtube.com/vi/gvjhrxwX7S8/0.jpg" alt="Introduction to
Amalgam8 Microservice Routing Fabric" width="240" height="180" border="10"
/></a>

2. [Try the demo application](examples/) in an container runtime of your
choice (Docker, Kubernetes, Marathon).

3. [Integrate the sidecar](https://www.amalgam8.io/docs/content/getting-started-with-amalgam8.html) into your existing apps and start using Amalgam8!

## What is Amalgam8 ?

Amalgam8 is a platform for building polyglot microservice applications that
enables you to route requests between microservices in a *content-based*
and *version-based* manner, independent of the underlying container
orchestration layer
([Docker Swarm](https://www.docker.com/products/docker-swarm),
[Kubernetes](https://kubernetes.io),
[Marathon](https://mesosphere.github.io/marathon/)) or the cloud platform
(Amazon AWS, IBM Bluemix, Google Cloud Platform, Microsoft Azure, etc.

Amalgam8 uses the sidecar model or the Ambassador pattern for building
microservices applications. The sidecar runs as in independent process and
takes care of service registration, discovery and request routing to
various microservices. The sidecar model simplifies development of polyglot
applications.

The Amalgam8 control plane dynamically programs the sidecars in each
microservice and provides you with a single pane of glass through which you
can control how requests are routed between microservices. Using the
control plane API, you can easily build tools for various DevOps tasks such
as simple A/B testing, internal releases and dark launches, canary
rollouts, red/black deployments, resilience testing, etc.


## Is this some new fangled stuff for microservices?

Absolutely not. In fact, this is exactly what you need in order to conduct
all sorts of smart experiments to measure user feedback, minimize risk
through smart rollout strategies, test your application's response to
failures, etc. To use more familiar terms, we are talking about simple A/B
testing, internal releases and dark launches, canary rollouts, red/black
deployments, and so on. Isn't this why you moved to the microservices
architecture in the first place?

Based on our experiences deploying and operating microservices inside IBM
at various scales, we realized that there are plenty of really cool 
tools out there that simplify service discovery and load balancing,
configuration management, secret storage, etc. There are even solutions for
API management for services exposed to end users, for tasks such as rate
limiting, authentication, subscription, etc.

But none of them were really addressing our core needs: **control how
microservices talk to each other**, so that we can quickly experiment
with different features, safely rollout new versions and test our
microservices in production confidently without worrying about bringing
down the entire infrastructure. We ended up writing automated tools that
achieved these tasks through clunky scripts, ugly DNS tricks, meddling with
the autoscaler just to achieve the right distribution of traffic between
the new and the old version.


## What is content and version-based routing? Why do I need it?

In any realistic production deployment, there are typically multiple
versions of microservices running at the same time, as you might be testing
out a new version, trouble shooting an old version, or simply keeping the
old version around just in case.

*Content-based routing* allows you to route requests between microservices
based on the content of the request, such as the URL, HTTP headers,
etc. This is very similar to the `location` blocks in Nginx, except more
powerful. For example, imagine being able to do something like this:

```
from microservice A, if request has "X-User-Id: QA", route to instance of
(B:v2) else route to instance of (B:v1)
```

*Version-based routing* allows you to control how different versions of
microservices can talk to each other. For example,

```
from microservice A:v2 route all requests to B:v2
from microservice A:v1 route 10% of requests to B:v2 and 90% to B:v1
```

You can program such rules for routing traffic between microservices using
the Amalgam8 Control Plane API.

---

## Structure of the repo

* The [examples](examples/) folder has some cool demo apps with detailed
  instructions to set up Amalgam8 in an container environment of your
  choice.

* Swagger documentation for the route controller can be found
  [here](controller/swagger.json). You can use these APIs to quickly build
  tools for canary deployments, failure recovery testing, etc.

* Swagger documentation for the service registry can be found
  [here](registry/swagger.json)

* The [devel](devel/) folder has the developer guide for those who wish to
  contribute features and bug fixes to Amalgam8
