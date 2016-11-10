Amalgam8 Command Line Interface (CLI)
=====================================

Command-line tool for Amalgam8 (http://www.amalgam8.io/).

Compressed binaries
-------------------

```
See the binaries folder.
```

For usage information, run the following command:
```
$ a8cli -h
```

Commands
--------
- a8cli rule-create
- a8cli rule-get
- a8cli rule-delete
- a8cli service-list
- a8cli route-list
- a8cli action-list
- a8cli info

Examples
--------

#### rule-create
* $ a8cli rule-create
```
*** Write .json or .yaml in a new line when finished ***
Enter DSL Rules:

rules:
- id: yaml_id
  destination: yaml_destination
  route:
    backends:
      - name: service1
        tags: [ v11, v12 ]
      - name: service2
        tags: [ v21, v22, yaml ]
.yaml
```
* $ a8cli rule-create -f rules.json

#### rule-get
* $ a8cli rule-get -a -o yaml
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
* $ a8cli rule-get -i f5f084aa-f813-4c94-b2a3-036c8779e5ed -o json
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
* $ a8cli rule-delete -i f5f084aa-f813-4c94-b2a3-036c8779e5ed
* $ a8cli rule-delete -d service1

#### service-list
  *  $ a8cli service-list
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

  * $ a8cli service-list -o yaml
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
* $ a8cli route-list
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
* $ a8cli action-list
```
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
| Destination | Rule Id                              | Priority | Match                                                   | Actions (EXPERIMENTAL)            |
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
| details     | 9c7198d7-d037-4cb6-8d48-b573608c7de9 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v1, prob=0,... |
| productpage | 0f12b977-9ab9-4d69-8dfe-3eae07c8f115 | 10       | source="gateway, header="Cookie:.*?user=jason"          | action=trace, tags=v1, prob=0,... |
| ratings     | c2d98e32-8fd0-4e0d-a363-8adff99b0692 | 10       | source="reviews:v2", header="Cookie:.*?user=jason"      | action=delay, tags=v1, prob=1,... |
| ratings     | 454a8fb0-d260-4832-8007-5b5344c03c1f | 10       | source="reviews:v2", header="Cookie:.*?user=jason"      | action=trace, tags=v1, prob=0,... |
| reviews     | 2d381a94-1796-45c3-a1d8-3965051b61b1 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v2, prob=0,... |
+-------------+--------------------------------------+----------+---------------------------------------------------------+-----------------------------------+
```

#### info
* $ a8cli info

```
Amalgam8 info...

Registry URL: http://127.0.0.1:31300
Registry Token: x3Eae07454aA8fb0c8f115

Controller URL: http://127.0.0.1:31200
Controller Token: 1796Ba1p573608c7de9

```
