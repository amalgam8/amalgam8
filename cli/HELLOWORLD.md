CLI command equivalents for Hello World example
============================================

#### Version-based routing
* Old
```
a8ctl route-set helloworld --default v1
```

* New
```
cat << EOF | a8ctl-beta rule-create -r
rules:
- priority: 1
  destination: helloworld
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
a8ctl route-set helloworld --default v1 --selector 'v2(weight=0.25)'
```

* New

  First, use the `rule-delete` command to delete all rules
```
a8ctl-beta rule-delete -a -f
```
  Create a new rule
```
cat << EOF | a8ctl-beta rule-create -r
rules:
- priority: 1
  destination: helloworld
  route:
    backends:
    - tags:
      - v2
      weight: 0.25
    - tags:
      - v1
EOF
```
  NOTE: It's also possible to get the same results by using `rule-get` to get the rule ID and `rule-update`


Documentation
--------

[Amalgam8 Rules DSL](https://www.amalgam8.io/docs/control-plane/controller/rules-dsl/)
