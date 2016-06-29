This is work in progress to provide a dynamic update framework for 
Amalgam8's nginx proxy, where nginx reload can be avoided. These updates 
are yet to be integrated with the rest of Amalgam8 code base.

To try the setup,
- start nginx

```bash
nginx -c ~/amalgam8/sidecar/dynamicupdate/nginx.conf
```

- start three versions of reviews service and one version of ratings service

```bash
nohup python reviews1.py 9081 dummy &
nohup python reviews2.py 9082 dummy &
nohup python reviews3.py 9083 dummy &
nohup python ratings1.py 9084 dummy &
```

- Populate the nginx proxy with reviews service, upstream and version selection info using the exposed API

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @config-3versions.json
```

- Now access reviews service. You should see that the output will alternative between three versions of reviews service.

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

and 

```
{
  "Reviewer4": 0,
  "Reviewer5": 0
}
```

depending on the version selector specified in the config-3versions.json file.

- Now, remove one of the reviews versions.

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @config-2versions.json
```

and access the reviews service as before. You will see the output
alternating between 2 versions of reviews service.

- Add a completely new service (ratings):

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @config-2services.json
```

and access the ratings service via

```bash
curl http://localhost:8888/ratings/ratings
```

- Remove ratings service completely

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @config-2versions.json
```

and try to access ratings service. You should see an error.


* TODO:
  - Add support for Gremlin rules in config.json
  - Add support for returning 404s when an upstream/service does not exist
  - Measure overhead and optimize
  - Modify sidecar receive config updates from controller and update nginx via API
  - Modify controller to generate config.json instead of Go template gen for nginx config
