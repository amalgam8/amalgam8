CLI command equivalents for Bookinfo example
============================================

#### List the Services in the App
* Old
```
a8ctl service-list
```

* New
```
a8ctl-beta service-list
```

#### Set the default routes
* Old
```
a8ctl route-set productpage --default v1
a8ctl route-set ratings --default v1
a8ctl route-set details --default v1
a8ctl route-set reviews --default v1
```

* New
```
cat << EOF | a8ctl-beta rule-create -r
rules:
- priority: 1
  destination: details
  route:
    backends:
    - tags:
      - v1
- priority: 1
  destination: ratings
  route:
    backends:
    - tags:
      - v1
- priority: 1
  destination: productpage
  route:
    backends:
    - tags:
      - v1
- priority: 1
  destination: reviews
  route:
    backends:
    - tags:
      - v1
EOF
```

#### List all routes
* Old
```
a8ctl route-list
```

* New
```
a8ctl-beta route-list
```

#### Content-based routing
* Old
```
a8ctl route-set reviews --default v1 --selector 'v2(user="jason")'
```

* New
```
cat << EOF | a8ctl-beta rule-create -r
rules:
- priority: 2
  destination: reviews
  match:
    headers:
      Cookie: .*?user=jason
  route:
    backends:
    - tags:
      - v2
EOF
```

#### Fault Injection w/ Manual Verification
* Old
```
a8ctl action-add --source reviews:v2 --destination ratings --cookie user=jason --action 'v1(1->delay=7)'
```

* New
```
cat << EOF | a8ctl-beta rule-create -r
rules:
- priority: 10
  destination: ratings
  match:
    source:
      name: reviews
      tags:
      - v2
    headers:
      Cookie: .*?user=jason
  actions:
  - action: delay
    duration: 7
    probability: 1
    tags:
    - v1
EOF
```

#### Fault Injection + Automated Verification
* Old
```
a8ctl rule-clear
```

* New

  First, use the `action-list` command to list all actions (`rule-get -a` could also be used)
```
a8ctl-beta action-list
```
  Copy the rule ID and delete it
```
a8ctl-beta rule-delete -i xxxxxxxxxxx
```

#### Run the gremlin recipe
* Old
```
a8ctl recipe-run --topology examples/bookinfo-topology.json --scenarios examples/bookinfo-gremlins.json --checks examples/bookinfo-checks.json --header 'Cookie' --pattern='user=jason'
```

* New
```
a8ctl-beta recipe-run -t examples/bookinfo-topology.json -s examples/bookinfo-gremlins.json -c examples/bookinfo-checks.json -H Cookie -p user=jason
```

#### Gradually migrate traffic to reviews:v3 for all users
* Old
```
a8ctl traffic-start reviews v3
a8ctl traffic-step reviews
a8ctl traffic-step reviews --amount 50
a8ctl traffic-step reviews --amount 100
```

* New
```
a8ctl-beta traffic-start -s reviews -v v3
a8ctl-beta traffic-step -s reviews
a8ctl-beta traffic-step -s reviews -a 50
a8ctl-beta traffic-step -s reviews -a 100
```

Documentation
--------

[Amalgam8 Rules DSL](https://www.amalgam8.io/docs/control-plane/controller/rules-dsl/)
