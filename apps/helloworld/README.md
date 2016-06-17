# Amalgam8 helloworld sample

## Overview

The helloworld sample starts two versions of a helloworld microservice, to demonstrate how Amalgam8 can be used to split 
incoming traffic between the two versions. You can define the proportion of traffic to each microservice as a percentage.

## Running the hellowrold demo

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

1. Confirm that the microservices are running, by running the following command:

    ```bash
    a8ctl service-list
    ```
    
    The expected output is the following:
    
    ```
    +------------+--------------+
    | Service    | Instances    |
    +------------+--------------+
    | helloworld | v1(2), v2(2) |
    +------------+--------------+
    ```

    There are 4 instances of the helloworld service. 2 are instances of version "v1" and 2 are version "v2". 

1. Send all traffic to the v1 version of helloworld, by running the following command:

    ```
    a8ctl route-set helloworld --default v1
    ```

1. You can confirm the routes are set by running the following command:

    ```bash
    a8ctl route-list
    ```

    You should see the following output:

    ```
    +------------+-----------------+-------------------+
    | Service    | Default Version | Version Selectors |
    +------------+-----------------+-------------------+
    | helloworld | v1              |                   |
    +------------+-----------------+-------------------+
    ```

1. Confirm that all traffic is being directed to the v1 instance, by running the following cURL command multiple times:

    ```
    curl ${GATEWAY_URL}/helloworld/hello
    ```

    **Note**: Replace GATEWAY_URL above with the appropriate URL of the gateway
    for your environment (for example, http://localhost:32000, http://192.168.33.33:32000, etc.).

    You can see that the traffic is continually routed between the v1 instances only, in a round-robin fashion:

    ```
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    ...
    ```

1. Next, we will split traffic between helloworld v1 and v2

    Run the following command to send 25% of the traffic to helloworld v2, leaving the rest (75%) on v1:
    
    ```
    a8ctl route-set helloworld --default v1 --selector 'v2(weight=0.25)'
    ```

1. Run this cURL command several times:

    ```
    curl ${GATEWAY_URL}/helloworld/hello
    ```

    You will see alternating responses from all 4 helloworld instances, where approximately 1 out of every 4 (25%) responses
    will be from a "v2" instances, and the other responses from the "v1" instances:

    ```
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v2, container: helloworld-v2-ggkvd
    $ curl ${GATEWAY_URL}/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    ...
    ```

    Note: if you use a browser instead of cURL to access the service and continually refresh the page, 
    it will always return the same version (v1 or v2), because a cookie is set to maintain version affinity.
    However, the browser still round-robins between the specific version instances that it returns.

## Using the Amalgam8 REST API

You can look at registration details for a service in the A8 registry, by running the following cURL command:

```
export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY
curl -X GET -H "Authorization: Bearer ${TOKEN}" ${REGISTRY_URL}/api/v1/services/helloworld | jq .
```

**Note**: Replace REGISTRY_URL above with the appropriate URL of the A8 Controller
for your environment (for example, http://localhost:31300, http://192.168.33.33:31300, etc.).

The output should look something like this:

```
{
  "instances": [
    {
      "last_heartbeat": "2016-04-27T20:43:36.306968276Z",
      "metadata": {
        "version": "v2"
      },
      "ttl": 45,
      "endpoint": {
        "value": "172.17.0.7:5000",
        "type": "http"
      },
      "service_name": "helloworld",
      "id": "a594b578955aa580"
    },
    {
      "last_heartbeat": "2016-04-27T20:43:36.610720426Z",
      "metadata": {
        "version": "v1"
      },
      "ttl": 45,
      "endpoint": {
        "value": "172.17.0.4:5000",
        "type": "http"
      },
      "service_name": "helloworld",
      "id": "9eec2aac0c6308f5"
    },
    {
      "last_heartbeat": "2016-04-27T20:43:36.673541582Z",
      "metadata": {
        "version": "v1"
      },
      "ttl": 45,
      "endpoint": {
        "value": "172.17.0.6:5000",
        "type": "http"
      },
      "service_name": "helloworld",
      "id": "69ce12035f9ada47"
    },
    {
      "last_heartbeat": "2016-04-27T20:43:36.718637643Z",
      "metadata": {
        "version": "v2"
      },
      "ttl": 45,
      "endpoint": {
        "value": "172.17.0.5:5000",
        "type": "http"
      },
      "service_name": "helloworld",
      "id": "161c6daaca4b23eb"
    }
  ],
  "service_name": "helloworld"
}
```

To list the routes for a service, run the following cURL command:

```
curl ${CONTROLLER_URL}/v1/tenants/local/versions/helloworld | jq .
```

**Note**: Replace CONTROLLER_URL above with the appropriate URL of the A8 Controller
for your environment (for example, http://localhost:31200, http://192.168.33.33:31200, etc.).

After running the demo, the output should be as follows:

```
{
  "selectors": "{v2={weight=0.25}}",
  "default": "v1",
  "service": "helloworld"
}
```

You can also set routes using the REST API. For example, to send all traffic to v2, run the following curl command:

```
curl -X PUT ${CONTROLLER_URL}/v1/tenants/local/versions/helloworld -d '{"default": "v2"}' -H "Content-Type: application/json"
```
