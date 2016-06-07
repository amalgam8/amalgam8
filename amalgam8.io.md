# Amalgam8 Microservices Fabric

## Overview

Amalgam8 is a multi-tenanted, platform and runtime agnostic, microservices integration framework.
The microservices in an Amalgam8 application can be polyglot, run in containers, VMs, or even on bare metal.

Amalgam8 provides proxy and registration sidecars and other adapter mechanisms that are used for registering and calling
microservices. Depending on a microservice's deployment runtime, integration with Amalgam8 can require little or no significant change
to the microservice impelementation code because Amalgam8 is designed to leverage the benefits of any given runtime
and then provide the easiest possible integration for each. For example, in Kubernetes, a microservice that adready has an associated
Kubernetes service definition can be registered automatically using an Amalgam8 plugin providing
an adapted view of the Kubernetes services in Amalgam8. 
Similar plugins are possible for other runtime environments, but where they are not, Amalgam8 also provides registration
and proxy functionality using sidecars in a number of predefined, but extensible, container images.

In addition to ease of integration, Amalgam8 also provides the integrated microservice-based applications with core functionality
enabling a great deal of control and testability, addressing some of the most challenging issues faced when moving to
a microservices-based system. This includes edge and mid-tier version-routing based on traffic percentage, user, and other
criteria. Delays and failures can also be injected into the path of calls to and between microservices, enabling advanced
end-to-end resiliency testing. And most importantly, all of this is designed for extensibility and is completely available in open source.

## How it Works

![high-level architecture](https://github.com/amalgam8/examples/blob/master/architecture.jpg)

At the heart of Amalgam8, are two mutli-tenanted services:

1. **Registry** - A high performance service registry that provides a centralized view of all the microservices in an application, regardless
   of where they are actually running.
2. **Controller** - Monitors the Registry and provides a REST API for registering routing and other microservice control-rules, which
   it uses to generate and send control information to proxy servers running within the application.

Application run as tenants of these two servers. They register their services in the Registry and use the Controller to manage proxies,
usually running in sidecars of the microservices.

![how it works](https://github.com/amalgam8/examples/blob/master/how-it-works.jpg)

1. Microservice instances are registered in the A8 Registry. There are several ways this may be accomplished (see below).
2. Administrator specifies routing rules and filters (e.g., version rules, test delays) to control traffic flow between microservices.
3. A8 Controller monitors the A8 Registry and administrator input and then generates control information that is sent to the A8 Proxies.
4. Requests to microservices are via an A8 Proxy (usually a client-side sidecar of another microservice)
5. A8 Proxy forwards request to an approriate microservice, depending on the request path and headers and the configuration specified by the controller.

## Tenant Library

Almagam8 includes a library containing a very flexible sidecar architecture that can be configured and used by tenants in a number of ways.

![sidecar architecture](https://github.com/amalgam8/examples/blob/master/sidecar.jpg)

...
