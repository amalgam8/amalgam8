# Amalgam8 - Microservice Routing Fabric

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/amalgam8
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/amalgam8
[Travis]: https://travis-ci.org/amalgam8/amalgam8
[Travis Widget]: https://travis-ci.org/amalgam8/amalgam8.svg?branch=master

## TL;DR

1. A quick intro video to Amalgam8

 <a href="http://www.youtube.com/watch?feature=player_embedded&v=gvjhrxwX7S8" target="_blank"><img
src="http://img.youtube.com/vi/gvjhrxwX7S8/0.jpg" alt="Introduction to
Amalgam8 Microservice Routing Fabric" width="240" height="180" border="10"/></a>

1. [Try the demo applications](https://amalgam8.io/_docs/demo/) with a
   container runtime of your choice.

1. [Integrate the sidecar](https://www.amalgam8.io/docs/sidecar/)
   into your existing application to start using Amalgam8.
   
---

## Content and version-based routing - 101

In any realistic production deployment, there are typically multiple
versions of microservices running at the same time, as you might be testing
out a new version, troubleshooting an old version, or simply keeping the
old version around just in case.

*Content-based routing* allows you to route requests between microservices
based on the content of the request, such as the URL, HTTP headers,
etc. For example,

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

*A simple way to accomplish these functions is to control how
microservices can talk to each other.*

# What is Amalgam8 ?

Amalgam8 is a platform for building polyglot microservice applications that
enables you to route requests between microservices in a *content-based*
and *version-based* manner, independent of the underlying container
orchestration layer
([Docker Swarm](https://www.docker.com/products/docker-swarm),
[Kubernetes](https://kubernetes.io),
[Marathon](https://mesosphere.github.io/marathon/)) or the cloud platform
(Amazon AWS, IBM Bluemix, Google Cloud Platform, Microsoft Azure, etc.)

Amalgam8 uses the sidecar model or the ambassador pattern for building
microservices applications. The sidecar runs as in independent process and
takes care of service registration, discovery and request routing to
various microservices. The sidecar model simplifies development of polyglot
applications.

Through the Amalgam8 Control Plane, you can dynamically program the
sidecars in each microservice and control how requests are routed between
microservices. The control plane provides REST APIs that serve as the basis
for building tools for various DevOps tasks such as A/B testing, internal
releases and dark launches, canary rollouts, red/black deployments,
resilience testing, etc.


## Amalgam8 - Components

* The Amalgam8 Control Plane consists of two multi-tenant components:
    * [Service Registry](https://amalgam8.io/docs/registry/)
    * [Route Controller](https://amalgam8.io/docs/controller/)

    The registry and the controller store their state in a Redis backend.

* In the data plane, the [Amalgam8 sidecar](https://amalgam8.io/docs/sidecar/) runs alongside each
  microservice instance. The sidecar is an
  [Nginx/OpenResty](https://openresty.org) reverse proxy. In addition to proxying
  requests to other microservices, the sidecar is responsible for service
  registration, heartbeat, service discovery, load balancing, intelligent
  request routing, and fault injection.
  
  Microservices communicate with the sidecar via the loopback socket at
  http://localhost:6379 . For e.g., to make a REST API call over HTTP to
  serviceB, the application would use the following URL:
  http://localhost:6379/serviceB/apiEndpoint . The sidecar in-turn forwards
  the API call to an instance of service B.

## Documentation

Detailed documentation on Amalgam8 can be found at [https://amalgam8.io/docs](https://amalgam8.io/docs). 

## Demos

To get started with Amalgam8, we suggest exploring some of the [demo
applications](https://amalgam8.io/docs/demo). The walkthroughs 
demonstrate some of Amalgam8's key features. Detailed instructions are
available for different container runtimes and cloud platforms.

## Getting Help

If you have any questions or feedback, you can reach us via our public
Slack channel (#amalgam8). To join this channel, please use the following
self invite URL: https://amalgam8-slack-invite.mybluemix.net

---

# Development Process

To build from source, clone this repository, and follow the instructions in
the [developer guide](devel/).

## Travis CI

Continuous builds are run on Travis CI. These builds use the `.travis.yml` configuration.


## Release Workflow

This section includes instructions for working with releases, and is intended for the project's maintainers (requires write permissions)

### Creating a release

1.  Edit the `CHANGELOG.md` file, describing the changes included in this release.

1.  Set a version for the release, by incrementing the current version
    according to the [semantic versioning](https://semver.org/)
    guidelines. For example,

    ```bash
    export VERSION=v0.1.0
    ```

1.  Create an [annotated tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging#Annotated-Tags) in your local copy of the repository:
   
    ```bash
    git tag -a -m "Release $VERSION" $VERSION [commit id]
    ```

    The `[commit id]` argument is optional. If not specified, HEAD is used.
   
1.  Push the tag back to the Amalgam8 upstream repository on GitHub:

    ```bash
    git push origin $VERSION
    ```
   This command automatically creates a release object on GitHub, corresponding to the pushed tag.
   The release contains downloadable packages of the source code (both as `.zip` and `.tag.gz` archives).

1.  Edit the [GitHub release object](https://github.com/amalgam8/amalgam8/releases), and add a title and description (according to `CHANGELOG.md`).


# License

Copyright 2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


# Contributing

Contributions and feedback are welcome! 
Proposals and pull requests will be considered. 
Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.
