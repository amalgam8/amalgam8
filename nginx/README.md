This is work in progress to provide a dynamic update framework for 
Amalgam8's nginx proxy, where nginx reload can be avoided. These updates 
are yet to be integrated with the rest of Amalgam8 code base.

To try the setup, first start nginx

```bash
cd $GOPATH/src/github.com/amalgam8/sidecar/dynamicupdate
nginx -c $GOPATH/src/github.com/sidecar/dynamicupdate/nginx.conf
```

This version of nginx also starts two versions of `helloworld` service
(each with 2 instances each) and one version of `endworld` service with
one instance. 

Lets see how we can update nginx dynamically to pick up new instances of
a service, new versions of a service and completely new services, without
having to reload nginx config. Finally, we will also setup failures in a
similar manner.

**Note**: The following commands should be run in the
`$GOPATH/src/github.com/amalgam8/sidecar/dynamicupdate` folder.

- Add new service `helloworld` (version 1)

```bash
$ curl -X POST http://localhost:8080/a8-admin -d @1_oneservice.json
```

Test if service is up:

```bash
$ curl http://localhost:8888/helloworld/
```

You should get `Helloworld v1 - instance 1` as the output.

- Add upstream servers for `helloworld_v1`

```bash
$ curl -X POST http://localhost:8080/a8-admin -d @2_oneservice_scaleout.json
```

Test the service:

```bash
$ curl http://localhost:8888/helloworld/
Helloworld v1 - instance 1
$ curl http://localhost:8888/helloworld/
Helloworld v1 - instance 2
$ curl http://localhost:8888/helloworld/
Helloworld v1 - instance 1
```

- Add a new version of service (`helloworld_v2`)

```bash
$ curl -X POST http://localhost:8080/a8-admin -d @3_oneservice_addversion.json
```

Test the service. The output should alternate accross service instances and
versions.

```bash
$ curl http://localhost:8888/helloworld/
Helloworld v1 - instance 1
$ curl http://localhost:8888/helloworld/
Helloworld v2 - instance 2
$ curl http://localhost:8888/helloworld/
Helloworld v2 - instance 1
$ curl http://localhost:8888/helloworld/
Helloworld v1 - instance 2
```

- Remove a version of a service (`helloworld_v1`) and scale down
  `helloworld_v2` to one instance.
  
```bash
$ curl -X POST http://localhost:8080/a8-admin -d @4_remove_version_instance.json
```

Test the service. You should only see output from one instance of
`helloworld_v2`.

```bash
$ curl http://localhost:8888/helloworld/
Helloworld v2 - instance 1
$ curl http://localhost:8888/helloworld/
Helloworld v2 - instance 1
```

- Add a new service called `endworld` (version 1) to the proxy. 

```bash
$ curl -v -X POST http://localhost:8080/a8-admin -d @5_add_new_service.json
```

Test the services.

```bash
$ curl http://localhost:8888/endworld/
End of world v1 - instance 1
$ curl http://localhost:8888/helloworld/
Helloworld v2 - instance 1
```

- Lets get rid of `helloworld` service completely.

```bash
$ curl -v -X POST http://localhost:8080/a8-admin -d @6_delete_service.json
```

Lets see if `helloworld` is still accessible.

```bash
$ curl http://localhost:8888/helloworld/
<html>
<head><title>404 Not Found</title></head>
<body bgcolor="white">
<center><h1>404 Not Found</h1></center>
<hr><center>openresty/1.9.15.1</center>
</body>
</html>
```

We get a HTTP 404, indicating that `helloworld` service has been deleted
from the proxy's routing tables.

- Lets inject failures into the API calls

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @7_inject_faults.json
```

This file sets up two faults: from the source service (represented by
nginx proxy as `source_v1`) and the destination services (`helloworld_v2` and
`endworld_v1`). Even though the JSON file has another entry for
`source_v2`, it will be ignored because those faults do not pertain to our
service `source_v1`.  In other words, the failures are specific to the
source_version and destination_version pair.

Lets see if `helloworld_v2` is still accessible from `source_v1` (which is
the current name of our nginx proxy). Note that we need to tag the requests
with a special header `X-Gremlin-Id` in order to trigger the fault injection.

```bash
$ time curl -H "X-Gremlin-Id: test-123" http://localhost:8888/helloworld/
Helloworld v1 - instance 1

real	0m0.014s
user	0m0.003s
sys	0m0.005s
$ time curl -H "X-Gremlin-Id: test-123" http://localhost:8888/helloworld/
Helloworld v2 - instance 1

real	0m3.016s
user	0m0.004s
sys	0m0.004s
```

Notice that `helloworld_v2` takes 3s to respond -- the delay we specified
in the JSON file. If you attempt to access `endworld_v1` service, you will
get a HTTP 503 error.

Now, lets imagine that the service being proxied by nginx has been upgraded
to `source_v2`. Kill nginx. Edit the config file and change the `server_name`
variable to `source_v2` under the `AMALGAM8 API` block and the `APP proxy`
block. Restart nginx and lets set up the failures:

```bash
curl -v -X POST http://localhost:8080/a8-admin -d @7_inject_faults.json
```

In this case, faults are injected only between `source_v2` to
`endworld_v2`. So, if you access helloworld, you should see no failures.

```bash
$ time curl -H "X-Gremlin-Id: test-123" http://localhost:8888/helloworld/
Helloworld v1 - instance 1

real	0m0.014s
user	0m0.003s
sys	0m0.005s
$ time curl -H "X-Gremlin-Id: test-123" http://localhost:8888/helloworld/
Helloworld v2 - instance 1

real	0m0.015s
user	0m0.004s
sys	0m0.004s
```

But if you access the endworld service, you should see an abrupt connection
termination, corresponding to a crash failure.

```bash
$ curl -H "X-Gremlin-Id: endtest2-123" http://localhost:8888/endworld/
curl: (52) Empty reply from server
```

---

* TODO:
  - Cleanup Lua code
  - **Measure overhead at high load and optimize lua code**
  - Modify sidecar to receive config updates from controller and update nginx via API
  - Modify controller to generate json payload instead of nginx config
