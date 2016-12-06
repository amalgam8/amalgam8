Amalgam8 Command Line Interface (CLI)
=====================================

Command-line tool for Amalgam8 (http://www.amalgam8.io/).

Build binary
-------------------

```
$ make build.cli
```
* See the bin folder.


For usage information, run the following command:
```
$ a8ctl-beta -h
```

Commands
--------
- a8ctl-beta service-list
- a8ctl-beta route-list
- a8ctl-beta action-list
- a8ctl-beta rule-create
- a8ctl-beta rule-update
- a8ctl-beta rule-get
- a8ctl-beta rule-delete
- a8ctl-beta traffic-start
- a8ctl-beta traffic-step
- a8ctl-beta traffic-abort
- a8ctl-beta recipe-run
- a8ctl-beta info

Examples
--------

#### rule-create
```
a8ctl-beta rule-create [ -f rules.yaml|rules.json ]
```
Create one or more routing or action rules described
by the [Rules DSL](https://www.amalgam8.io/docs/control-plane/controller/rules-dsl/)
in the specified JSON or YAML file. If the -f option is not specified, the Rules DSL is read from stdin.

* $ a8ctl-beta rule-create
```
#########################################################################
Press (CTRL+D) on Unix-like systems or (Ctrl+Z) in Windows when finished.
You can also write .json or .yaml in a new line to finish.
#########################################################################

Enter DSL Rules:

rules:
- destination: yaml_destination
  route:
    backends:
      - name: service1
        tags: [ v11, v12 ]
      - name: service2
        tags: [ v21, v22, yaml ]
.yaml
```
* $ a8ctl-beta rule-create -f rules.json

#### rule-update
```
a8ctl-beta rule-update [ -f rules.yaml|rules.json ]
```
Update one or more routing or action rules described
by the [Rules DSL](https://www.amalgam8.io/docs/control-plane/controller/rules-dsl/)
in the specified JSON or YAML file. If the -f option is not specified, the Rules DSL is read from stdin.

* $ a8ctl-beta rule-update
```
#########################################################################
Press (CTRL+D) on Unix-like systems or (Ctrl+Z) in Windows when finished.
You can also write .json or .yaml in a new line to finish.
#########################################################################

Enter DSL Rules:

rules:
- id: f5f084aa-f813-4c94-b2a3-036c8779e5ed
  destination: yaml_destination
  route:
    backends:
      - name: service1
        tags: [ v11, v12 ]
      - name: service2
        tags: [ v21, v22, beta ]
.yaml
```
* $ a8ctl-beta rule-update -f rules.json

#### rule-get
```
a8ctl-beta rule-get [-a] [-i id]* [-t tag]* [-d destination]* [-o json|yaml]
```
Output the [Rules DSL](https://www.amalgam8.io/docs/control-plane/controller/rules-dsl/)
of one or more rules with the specified rule ids, tags, or destinations.
If the -a option is specified, all defined rules are output.
Rule DSL is output in JSON by default but can be changed using the -o option.

* $ a8ctl-beta rule-get -a -o yaml
```
- id: f5f084aa-f813-4c94-b2a3-036c8779e5ed
  destination: yaml_destination
  route:
    backends:
    - name: service1
      tags:
      - v11
      - v12
    - name: service2
      tags:
      - v21
      - v22
      - yaml
- id: 454a8fb0-d260-4832-8007-5b5344c03c1f
  priority: 10
  tags:
  - 9f0ea878-a1f4-11e6-b410-6c40089c9f90
  destination: ratings
  match:
    source:
      name: reviews
      tags:
      - v2
    headers:
      Cookie: .*?user=jason
  actions:
  - action: trace
    tags:
    - v1
    log_key: gremlin_recipe_id
    log_value: 9f0ea878-a1f4-11e6-b410-6c40089c9f90
  - action: delay
    duration: 7
    probability: 1
    tags:
    - v1
```
* $ a8ctl-beta rule-get -i f5f084aa-f813-4c94-b2a3-036c8779e5ed -o json
```
{
  "rules": [
    {
      "id": "f5f084aa-f813-4c94-b2a3-036c8779e5ed",
      "destination": "yaml_destination",
      "route": {
        "backends": [
          {
            "name": "service1",
            "tags": [
              "v11",
              "v12"
            ]
          },
          {
            "name": "service2",
            "tags": [
              "v21",
              "v22",
              "yaml"
            ]
          }
        ]
      }
    }
  ]
}
```

#### rule-delete
```
a8ctl-beta rule-delete [-i id]* [-t tag]* [-d destination]* [-f]
```
Delete one or more rules with the specified rule ids, tags, or destinations.
If the -f option is specified, no confirmation will be asked.

* $ a8ctl-beta rule-delete -i f5f084aa-f813-4c94-b2a3-036c8779e5ed
* $ a8ctl-beta rule-delete -d service1
* $ a8ctl-beta rule-delete -a -f

#### service-list
```
a8ctl-beta service-list [-o json|yaml]
```
Output a table listing all of the currently defined services along with any active instances of them.

*  $ a8ctl-beta service-list
```
+-------------+---------------------+
| Service     | Instances           |
+-------------+---------------------+
| reviews     | v1(1), v2(1), v3(1) |
| productpage | v1(1)               |
| ratings     | v1(1)               |
| helloworld  | v1(2), v2(2)        |
| details     | v1(1)               |
+-------------+---------------------+
```

* $ a8ctl-beta service-list -o yaml
```
- service: details
    instances:
    - v1(1)
  - service: productpage
    instances:
    - v1(1)
  - service: reviews
    instances:
    - v3(1)
    - v2(1)
    - v1(1)
  - service: ratings
    instances:
    - v1(1)
```

#### route-list
```
a8ctl-beta route-list [-o json|yaml]
```
Output a table listing all of the currently defined routing-type rules.

* $ a8ctl-beta route-list
```
+------------------+-----------------+--------------------------------------+
| Service          | Default Version | Version Selectors                    |
+------------------+-----------------+--------------------------------------+
| service1         | v21,v22,json    |                                      |
| ratings          | v2              | v2(user="jason")                     |
| reviews          | v2              | v2(header="Foo:bar"), v3(weight=0.5) |
| details          |                 |                                      |
| productpage      |                 |                                      |
+------------------+-----------------+--------------------------------------+
```

#### action-list

```
a8ctl-beta action-list [-o json|yaml]
```
Output a table listing all of the currently defined action-type rules.

* $ a8ctl-beta action-list
```
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
| Destination | Rule Id                              | Priority | Match                                                   | Actions                           |
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
| details     | 9c7198d7-d037-4cb6-8d48-b573608c7de9 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v1, prob=0.75  |
| productpage | 0f12b977-9ab9-4d69-8dfe-3eae07c8f115 | 10       | source="gateway, header="Cookie:.*?user=jason"          | action=trace, tags=v1, prob=1.0   |
| ratings     | c2d98e32-8fd0-4e0d-a363-8adff99b0692 | 10       | source="reviews:v2", header="Cookie:.*?user=jason"      | action=delay, tags=v1, prob=1.0   |
| ratings     | 454a8fb0-d260-4832-8007-5b5344c03c1f | 10       | source="reviews:v2", header="Cookie:.*?user=jason"      | action=trace, tags=v1, prob=0.5   |
| reviews     | 2d381a94-1796-45c3-a1d8-3965051b61b1 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v2, prob=0.5   |
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
```

#### traffic-start
```
a8ctl-beta traffic-start [-s service] [-v version] [-a amount]
```
Start transferring traffic from one version of a microservice to another.
If the -a option is not specified, the default value will be used [-a 10].

* a8ctl-beta traffic-start -s reviews -v v3
```
Transfer starting for "reviews": diverting 10% of traffic from "v1" to "v3"
```

#### traffic-step
```
a8ctl-beta traffic-step [-s service] [-a amount]
```
Increase the amount of traffic sent to an specific version of a microservice.

* a8ctl-beta traffic-step -s reviews
```
Transfer step for reviews: diverting 20% of traffic from v1 to v3
```
* a8ctl-beta traffic-step -s reviews -a 90
```
Transfer step for reviews: diverting 90% of traffic from v1 to v3
```
* a8ctl-beta traffic-step -s reviews
```
Transfer complete for reviews: sending 100% of traffic to v3
```

#### traffic-abort
```
a8ctl-beta traffic-abort [-s service]
```
Stop transferring traffic to another version of a microservice.

* a8ctl-beta traffic-abort -s reviews
```
Transfer aborted for reviews: all traffic reverted to v1
```

#### recipe-run
```
a8ctl-beta recipe-run [-t topology] [-s scenarios] [-c checks] [-r run-load-script] [-H header] [-p pattern] [-w wait] [-f force] [-o output]
```
Run a gremlin recipe that describes the application topology, scenario and a set of assertions for fault injection and automated verification.

* a8ctl-beta recipe-run -t examples/bookinfo-topology.json -s examples/bookinfo-gremlins.json -c examples/bookinfo-checks.json -H Cookie -p user=jason


#### info
* $ a8ctl-beta info

```
Amalgam8 info...

+---------------------+------------------------+
| Env. Variable       | Value                  |
+---------------------+------------------------+
| A8_REGISTRY_URL     | http://127.0.0.1:31300 |
| A8_REGISTRY_TOKEN   |                        |
| A8_CONTROLLER_URL   | http://127.0.0.1:31200 |
| A8_CONTROLLER_TOKEN |                        |
| A8_GREMLIN_URL      | http://127.0.0.1:31500 |
| A8_GREMLIN_TOKEN    |                        |
| A8_DEBUG            | true                   |
+---------------------+------------------------+
```
