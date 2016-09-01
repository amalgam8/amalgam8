# Sidecar

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/sidecar
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/sidecar
[Travis]: https://travis-ci.org/amalgam8/sidecar
[Travis Widget]: https://travis-ci.org/amalgam8/sidecar.svg?branch=master

Ambassador pattern for microservices made simple and powerful. The Amalgam8
sidecar enables intelligent request routing while automating service
registration, discovery, and client-side load-balancing. The Amalgam8
sidecar is based on Go+Nginx and follows an architecture similar to
[Netflix Prana sidecar](http://techblog.netflix.com/2014/11/prana-sidecar-for-your-netflix-paas.html) 
or [AirBnB Smartstack](http://nerds.airbnb.com/smartstack-service-discovery-cloud/).


An overview of the Amalgam8 project is available here: https://amalgam8.io/

Documentation related to the sidecar can be found at https://amalgam8.io/docs

## TL;DR

* Install the sidecar in your Dockerized microservice.

    ```Dockerfile
    RUN curl -sSL https://git.io/a8sidecar.sh | sh
    ```

* Launch your app via the sidecar

    ```Dockerfile
    ENTRYPOINT ["a8sidecar", "--supervise", "YOURAPP", "YOURAPP_ARG", "YOURAPP_ARG"]
    ```

* Make API calls to other microservices via the sidecar

    [http://localhost:6379/\<serviceName\>/\<endpoint\>]()

* Control traffic to different versions of microservices using the
[a8ctl](https://github.com/amalgam8/a8ctl) utility

    ```bash
    a8ctl route-set serviceName --default v1 --selector 'v2(user="Alice")' --selector 'v3(user="Bob")'
    ```


## 1. Integrating the sidecar into your application

### Single Docker container <a id="int-docker"></a>

Add the following line to your `Dockerfile` to install the sidecar in your docker container:

```Dockerfile
RUN curl -sSL https://git.io/a8sidecar.sh | sh
```

or

```Dockerfile
RUN wget -qO- https://git.io/a8sidecar.sh | sh
```

The above URL points to the latest stable release of Amalgam8 sidecar. If
you would like to install a specific release, replace the URL with 
`https://github.com/amalgam8/sidecar/releases/download/${VERSION}/install-a8sidecar.sh`
where `${VERSION}` is the version of the sidecar that you wish to install.

**Optional app supervision:** The sidecar can serve as a supervisor process that
automatically starts up your application in addition to the Nginx proxy. To
use the sidecar to manage your application, add the following lines to your
`Dockerfile`

```Dockerfile
ENTRYPOINT ["a8sidecar", "--supervise", "YOURAPP", "YOURAPP_ARG", "YOURAPP_ARG"]
```

If you wish to manage the application process by yourself, then make sure
to launch the sidecar in the background when starting the docker
container. The environment variables required to run the sidecar are
described in detail [below](#runtime).

### Kubernetes Pods <a id="int-kube"></a>

With Kubernetes, the sidecar can be run as a standalone container in the
same `Pod` as your application container. No changes are needed to the
application's Dockerfile. Modify your service's YAML file to launch the
sidecar as another container in the same pod as your application
container. The latest version of the sidecar is available in Docker Hub in
two formats:

*  `amalgam8/a8-sidecar` - ubuntu-based version
*  `amalgam8/a8-sidecar:alpine` - alpine linux based version

## 2. Starting the sidecar <a id="runtime"></a>

The following instructions apply to both Docker-based and Kubernetes-based
installations. There are two modes for running the sidecar:

<!-- ### Environment variables or CLI flags -->

<!-- An exhaustive list of configuration options can be found in the -->
<!-- [Configuration](#config) section. For a quick start, take a look at the -->
<!-- [examples apps](https://github.com/amalgam8/examples) to get an idea of the -->
<!-- required environment variables needed by Amalgam8. -->


### With automatic service registration only <a id="regonly"></a>

For leaf nodes, i.e., microservices that make no outbound calls, only
service registration is required. Inject the following environment
variables while launching your application container in Docker or the
sidecar container inside kubernetes 

```bash
A8_PROXY=false
A8_REGISTER=true
A8_REGISTRY_URL=http://a8registryURL
A8_REGISTRY_TOKEN=a8registry_auth_token
A8_REGISTRY_POLL=polling_interval_between_sidecar_and_registry(5s)
A8_SERVICE=service_name:service_tags
A8_ENDPOINT_PORT=port_where_service_is_listening
A8_ENDPOINT_TYPE=http|https|tcp|udp|user
```

### With automatic service registration, discovery & intelligent routing <a id="routing"></a>

For microservices that make outbound calls to other microservices, service
registration, service discovery and client-side load balancing,
version-aware routing are required.

```bash
A8_REGISTER=true
A8_REGISTRY_URL=http://a8registryURL
A8_REGISTRY_TOKEN=a8registry_auth_token
A8_REGISTRY_POLL=polling_interval_between_sidecar_and_registry(5s)
A8_SERVICE=service_name:service_tags
A8_ENDPOINT_PORT=port_where_service_is_listening
A8_ENDPOINT_TYPE=http|https|tcp|udp|user

A8_PROXY=true
A8_LOG=false
A8_CONTROLLER_URL=http://a8controllerURL
A8_CONTROLLER_TOKEN=a8controller_auth_token
A8_CONTROLLER_POLL=polling_interval_between_sidecar_and_controller(5s)
```

**Update propagation**: The sidecar will periodically poll the Amalgam8
Controller for rule updates, and the Amalgam8 Registry for to obtain list
of registered instances of various microservices.

**Request logs**: All logs pertaining to external API calls made by
the Nginx proxy will be stored in `/var/log/nginx/a8_access.log` and
`/var/log/nginx/error.log`. The access logs are stored in JSON format. Note
that there is **no support for log rotation**. If you have a monitoring and
logging system in place, it is advisable to propagate the request logs to
your log storage system in order to take advantage of Amalgam8 features
like resilience testing.

The sidecar installation comes preconfigured with
[Filebeat](https://www.elastic.co/products/beats/filebeat) that can be
configured automatically to ship the access logs to a Logstash server,
which in turn propagates the logs to elasticsearch. If you wish to use the
filebeat system for log processing, make sure to have Elasticsearch and
Logstash services available in your application deployment. The following
two environment variables enable the filebeat process:

```bash
A8_LOG=true
A8_LOGSTASH_SERVER='logstash_server:port'
```

**Note:** The logstash environment variable needs to be enclosed in single quotes.

## 3. Using the sidecar

The sidecar is independent of your application process. The communication
model between a microservice, its sidecar and the target microservice is
shown below:

![Communication between app and sidecar](http://cdn.rawgit.com/amalgam8/sidecar/master/communication-model.svg)

When you want to make API calls to other microservices from your
application, you should call the sidecar at localhost:6379. 
The format of the API call is
[http://localhost:6379/\<serviceName\>/\<endpoint\>]()

where the `serviceName` is the service name that was used when launching
the target microservice (the `A8_SERVICE` environment variable), and the
endpoint is the API endpoint exposed by the target microservice.

For example, to invoke the `getItem` API in the `catalog` microservice,
your microservice would simply invoke the API via the URL:
[http://localhost:6379/catalog/getItem?id=123]().

Note that service versions are not part of the URL. The choice of the
service version (e.g., catalog:v1, catalog:v2, etc.), will be done
dynamically by the sidecar, based on routing rules set by the Amalgam8
controller.

---

## Configuration options <a id="config"></a>

Configuration options can be set via environment variables, command line flags, or YAML configuration files.
Order of precedence is command line flags first, then environmenmt variables, configuration files, and lastly default values.

**Note:** Atleast one of `A8_REGISTER` or `A8_PROXY` must be true.

| Environment Variable | Flag Name                   | YAML Key | Description | Default Value |Required|
|:---------------------|:----------------------------|:---------|:------------|:--------------|--------|
| A8_CONFIG | --config | | Path to a file to load configuration from | | no |
| A8_LOG_LEVEL | --log_level | log_level | Logging level (debug, info, warn, error, fatal, panic) | info | no |
| A8_SERVICE | --service | service.name & service.tags | service name to register with, optionally followed by a colon and a comma-separated list of tags | | yes |
| A8_ENDPOINT_HOST | --endpoint_host | endpoint.host | service endpoint hostname. Defaults to the IP (e.g., container) where the sidecar is running | optional |
| A8_ENDPOINT_PORT | --endpoint_port | endpoint.port | service endpoint port |  | yes |
| A8_ENDPOINT_TYPE | --endpoint_type | endpoint.type | service endpoint type (http, https, udp, tcp, user) | http | no |
| A8_REGISTER | --register | register | enable automatic service registration and heartbeat | false | See note above |
| A8_PROXY | --proxy | proxy | enable automatic service discovery and load balancing across services using NGINX | false | See note above |
| A8_LOG | --log | log | enable logging of outgoing requests through proxy using FileBeat | false | no |
| A8_SUPERVISE | --supervise | supervise | Manage application process. If application dies, sidecar process is killed as well. All arguments provided after the flags will be considered as part of the application invocation | false | no |
| A8_REGISTRY_URL | --registry_url | registry.url | registry URL |  | yes if `-register` is enabled |
| A8_REGISTRY_TOKEN | --registry_token | registry.token | registry auth token | | yes if `-register` is enabled and an auth mode is set |
| A8_REGISTRY_POLL | --registry_poll | registry.poll | interval for polling Registry | 15s | no |
| A8_CONTROLLER_URL | --controller_url | controller.url | controller URL |  | yes if `-proxy` is enabled |
| A8_CONTROLLER_TOKEN | --controller_token | controller.token | Auth token for Controller instance |  | yes if `-proxy` is enabled and an auth mode is set |
| A8_CONTROLLER_POLL | --controller_poll | controller.poll | interval for polling Controller | 15s | no |
| A8_LOGSTASH_SERVER | --logstash_server | logstash_server | logstash target for nginx logs |  | yes if `-log` is enabled |
|  | --help, -h | show help | | |
|  | --version, -v | print the version | | |

### Configuration precedence

### Example configuration file:
```yaml
register: true
proxy: true

service:
  name: helloworld
  tags: 
    - v1
    - somethingelse
  
endpoint:
  host: 172.10.10.1
  port: 9080
  type: https

registry:
  url:   http://registry:8080
  token: abcdef
  poll:  10s
  
controller:
  url:   http://controller:6379
  token: abcdef
  poll:  30s
  
supervise: true
app: [ "python", "helloworld.py ]

log: true
logstash_server: logstash:8092

log_level: debug
```
---

## Building from source

The following sections describe options for building the sidecar from
source.

### Preprequisites

* Docker 1.8 or higher
* Go 1.6

### Clone

Clone the repository manually, or use `go get`:

```go get github.com/amalgam8/sidecar```

### Make targets

The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `release`        | *(Default)* `release` builds the sidecar within a docker container and packages it into an image |
| `test`           | `test` runs all tests using `go test` |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

## Release Workflow

This section includes instructions for working with releases, and is intended for the project's maintainers (requires write permissions)

### Creating a release

1.  Set a version for the release, by incrementing the current version according to the [semantic versioning](https://semver.org/) guidelines:
   
    ```bash
    export VERSION=v0.1.0
    ```

1.  Update the APP_VER variable in the Makefile such that it matches with
    the VERSION variable above.

1.  Create an [annotated tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging#Annotated-Tags) in your local copy of the repository:
   
    ```bash
    git tag -a -m "Release $VERSION" $VERSION [commit id]
    ```

    The `[commit id]` argument is optional. If not specified, HEAD is used.
   
1.  Push the tag back to the Amalgam8 upstream repository on GitHub:

    ```bash
    git push upstream $VERSION
    ```
   This command automatically creates a release object on GitHub, corresponding to the pushed tag.
   The release contains downloadable packages of the source code (both as `.zip` and `.tag.gz` archives).

1.  Edit the `CHANGELOG.md` file, describing the changes included in this release.

1.  Edit the [GitHub release object](https://github.com/amalgam8/sidecar/releases), and add a title and description (according to `CHANGELOG.md`).

## License
Copyright 2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Contributing

Contributions and feedback are welcome!
Proposals and pull requests will be considered. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/amalgam8.github.io/blob/master/CONTRIBUTING.md)
file for more information.

