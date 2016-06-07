# Amalgam8 helloworld sample

## Overview

The helloworld sample starts two versions of a helloworld microservice, to demonstrate how Amalgam8 can be used to split the incoming traffic between the two versions. You can define the proportion of traffic to each microservice as a percentage.

## Starting the helloworld instances

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

1. Start the helloworld sample by running the following commands:
  ```
    cd $GOPATH/src/github.com/amalgam8/examples/apps/helloworld
    ./run.sh
  ```

2. After the helloworld instances are created, view their entries in the registry, by running the following cURL command:
  ```
    curl -X GET -H "Authorization: Bearer ${TOKEN}" http://${AR}/api/v1/services/helloworld | jq .
  ```

  There are 4 registered instances, 2 instances named "v1" and 2 that are named "v2". These represent the different versions of the helloworld instance. The output will resemble the following example:

  ```
    $ curl -X GET -H "Authorization: Bearer ${TOKEN}" http://${AR}/api/v1/services/helloworld | jq .
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

### Sending traffic to helloworld v1

3. Send all traffic to the v1 version of helloworld, by setting the rules with the following cURL command:

  ```
    curl -X PUT http://${AC}/v1/tenants/local/versions/helloworld -d '{"default": "v1"}' -H "Content-Type: application/json"
  ```

4. If you want to view the rules that are applied to helloworld v1, run the following cURL command:

  ```
    curl http://${AC}/v1/tenants/local/versions/helloworld | jq .
    {
      "selectors": "",
      "default": "v1",
      "service": "helloworld"
    }
  ```

5. Confirm that all traffic is being directed to the v1 instance, by running the following cURL command multiple times:

  ```
    curl 192.168.33.33:32000/helloworld/hello
  ```

  You can see that the traffic is continually routed between the v1 instances only, in a round-robin configuration:

  ```
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    ...
  ```

### Splitting traffic between helloworld v1 and v2

  Next, we will route some of the traffic to helloworld v1, and some to helloworld v2.

6. Run the following cURL command to change the rule to send 25% of the traffic to helloworld v2:

  ```
    curl -X PUT http://${AC}/v1/tenants/local/versions/helloworld -d '{"default": "v1", "selectors": "{v2={weight=0.25}}"}' -H "Content-Type: application/json"
  ```

7. Run this cURL command several times:

  ```
    curl 192.168.33.33:32000/helloworld/hello
  ```

  You will see alternating responses from all 4 helloworld instances, where approximately 1 out of every 4 (25%) responses will be from a "v2" instance, and the other responses from "v1":

  ```
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-qwpex
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v2, container: helloworld-v2-ggkvd
    $ curl 192.168.33.33:32000/helloworld/hello
    Hello version: v1, container: helloworld-v1-p8909
    ...
  ```

  Note: if you use a browser instead of cURL to access the service and continually refresh the page, 
  it will always return the same version (v1 or v2), because a cookie is set to maintain version affinity.
  However, the browser still round-robins between the specific version instances that it returns.

## Shutting down

8. To shutdown the helloworld instances, run the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/apps/helloworld
    kubectl delete -f helloworld.yaml
  ```
