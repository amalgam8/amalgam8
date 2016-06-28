This is work in progress to provide a dynamic update framework for 
Amalgam8's nginx proxy, where nginx reload can be avoided. These updates 
are yet to be integrated with the rest of Amalgam8 code base.

To try the setup,
* start nginx

```bash
nginx -c ~/amalgam8/sidecar/dynamicupdate/nginx.conf
```

* start two versions of reviews service

```bash
nohup python reviews1.py 9081 dummy &
nohup python reviews2.py 9082 dummy &
```

* Populate the nginx proxy with upstream info and version selection info using the exposed API

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @config.json
```

* Now access reviews service. You should see that the output will alternative between two versions of reviews service.

```bash
curl http://localhost:8888/reviews/reviews
```

Your output would alternate between

```
{
  "Baz": "zoo",
  "Foo": "bar"
}
```

and

```
{
  "Reviewer1": 5,
  "Reviewer2": 4
}
```

depending on the version selector specified in the config.json file.


* TODO:
  - Add support for Gremlin rules in config.json
  - Add support for dynamic creation of upstream blocks instead of hard coding them in the nginx.conf file
  - Measure overhead and optimize
  - Modify sidecar to generate initial nginx config and then, upon receiving config updates from controller, update nginx via API
  - Modify controller to generate config.json instead of Go template gen for nginx config. Move config generation to sidecar.
