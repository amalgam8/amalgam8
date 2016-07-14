This is work in progress to provide a dynamic update framework for 
Amalgam8's nginx proxy, where nginx reload can be avoided. These updates 
are yet to be integrated with the rest of Amalgam8 code base.

To try the setup:

```bash
docker build -f docker/Dockerfile.ubuntu -t nginx_trial .
docker run -it -e A8_SERVICE=source -e A8_SERVICE_VERSION=v1 --entrypoint /bin/bash -v `pwd`/nginx/lua:/opt/a8_lualib -v `pwd`/nginx/old:/opt/old nginx_trial
```

Once you are inside the container, start the main nginx server and the
upstream nginx server using the following commands:

```bash
nginx
nginx -c /opt/old/nginx.conf.plain
```

Now, initialize the upstreams with
```bash
curl -X POST http://localhost:5813/a8-admin -d @/opt/old/7_inject_faults.json
```

and test with

```bash
curl -H 'X-Gremlin-Id: test-124' http://localhost:6379/helloworld/
```

For helloworld-v2, you would get a 3s delay.
