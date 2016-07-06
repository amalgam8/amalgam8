# Amalgam8 Service Registry

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/registry
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/registry
[Travis]: https://travis-ci.org/amalgam8/registry
[Travis Widget]: https://travis-ci.org/amalgam8/registry.svg?branch=master

The Amalgam8 Service Registry is software that was developed to register and locate instances in applications 
that are built using a microservice-based architecture. The registry can be used to provide multi-tenant registration and 
discovery namespaces (i.e., isolation between different tenant scopes), a unified discovery for service
instances registered using different backends (e.g., Kubernetes, Eureka), and different authentication and authorization
backends. 

**Note**: The Amalgam8 Service Registry is currently in alpha. This means that the API will change, features will be added/removed, 
and that there will be bugs. 

## Building and Running

A recent prebuilt Amalgam8 Service Registry image can be found in docker.io/amalgam8/a8-registry:0.1.
```sh
$ docker pull amalgam8/a8-registry:0.1
```

If you wish to build the Amalgam8 Service Registry from source, clone this repository, and follow the instructions below.

### Preprequisites
A developer-tested development environment requires the following technologies:
* Linux host (tested with Ubuntu 16.04 LTS).
* Docker (tested with 1.10).
* Go toolchain (tested with 1.6.x). See [Go downloads](https://golang.org/dl/) and [installation instructions](https://golang.org/doc/install).


### Building a Docker Image

The Amalgam8 Service Registry can be built by simply typing `make build docker` with the [Docker
daemon](https://docs.docker.com/installation/) (v1.10.x) running.

This produces an image tagged `registry:0.1` that you can run.

### Standalone

The Amalgam8 Service Registry can also be run outside of a docker container as a Go binary. 
This is not recommended for production, but it can be useful for development or easier integration with 
your local Go tools.

The following commands will build and run it outside of Docker:

```
make build
./bin/registry
```

### Make Targets

The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `build`          | *(Default)* `build` builds the registry binary in the ./bin directory |
| `precommit`      | `precommit` should be run by developers before committing code changes. It runs code formatting and checks. |
| `docker`         | `docker` packages the binary in a docker container |
| `test`           | `test` runs (short duration) tests using `go test`. You may also `make test.all` to include long running tests. |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

### Continuous Integration with Travis CI

Continuous builds are run on Travis CI. These builds use the `.travis.yml` configuration.

## Usage

The Amalgam8 Service Registry supports a number of configuration options, most of which are set through environment variables.

The environment variables can be set via command line flags as well. 

### Command Line Flags and Environment Variables

The following environment variables are available. All of them are optional.

#### Registry Configuration

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| `API_PORT` | `--api_port` | API port number | 8080 |
| `LOG_LEVEL` | `--log_level` | Logging level. Supported values are: `debug`, `info`, `warn`, `error`, `fatal`, `panic` | `debug` |
| `LOG_FORMAT` | `--log_format` | Logging format. Supported values are: `text`, `json`, `logstash` | `text` |
| `NAMESPACE_CAPACITY` | `--namespace_capacity` | maximum number of instances that may be registered in a namespace | -1 (no capacity limit) |  
| `DEFAULT_TTL` | `--default_ttl` | Registry default instance time-to-live (TTL) | 30s |
| `MIN_TTL` | `--min_ttl` | Minimum TTL that may be specified during registration | 10s | 
| `MAX_TTL` | `--max_ttl` | Maximum TTL that may be specified during registration | 10m |


#### Authentication and Authorization

The Amalgam8 Service Registry optionally supports multi-tenancy by isolating each tenant into a separate namespace.
A namespace is defined by an opaque string carried in the HTTP `Authorization` header of API requests. The following
namespace authorization methods are supported and controlled via the `AUTH_MODE` environment variable (or `--auth_mode`
flag):
* None: if no authorization mode is defined, all instances are registered into a default shared namespace. 
* Trusted: namespace is retrieved directly from the Authorization header. This provides namespace separation in a trusted
environment (e.g., single tenant with multiple applications or environments).
* JWT: encodes the namespace value in a signed JWT token claim. 

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| `AUTH_MODE` | `--auth_mode` | Authentication modes. Supported values are: `trusted`, `jwt` | none (no isolation) |
| `JWT_SECRET` | `--jwt_secret` | Secret key for JWT authentication | none (must be set if `AUTH_MODE` is `jwt`) |
| `REQUIRE_HTTPS` | `--require_https` | Require clients to use HTTPS for API calls | `false` |

If `jwt` is specified, `JWT_SECRET` (or `--jwt_secret`) must be set as well to allow encryption and decryption.
Namespace value encoding must be present in every API call using HTTP Bearer Authorization:

```sh
Authorization: Bearer jwt.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ...wifQ.Gbz4G_O...NqdY`
```

#### Clustering

Amalgam8 Service Registry uses a memory only storage solution, without persistency (although different storage 
backends can be implemented). To provide HA and scale, the registry can be run in a cluster and supports replication
between cluster members.

Peer discovery currently uses a shared volume between all members. The volume must be mounted RW into each container.
Alternative discovery mechanisms are being explored.

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| `CLUSTER_SIZE` | `--cluster_size` | Cluster minimal healthy size, peers detecting a lower value will log errors | 1 (standalone) |
| `CLUSTER_DIR` | `--cluster_dir` | Filesystem directory for cluster membership | none, must be specified for clustering to work |
| `REPLICATION` | `--replication` | Enable replication between cluster members | `false` |
| `REPLICATION_PORT` | `--replication_port` | Replication port number | 6100 |
| `SYNC_TIMEOUT` | `--sync_timeout` | Timeout for establishing connections to peers for replication | 30s |

#### Catalog Extensions

The Amalgam8 Service Registry optionally supports read-only catalogs extensions.
The content of each catalog extension (e.g., Kubernetes, Docker-Swarm, etc) is read by the Amalgam8 Service Registry and
returned to the user along with the content of the registry itself.

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| `K8S_URL` | `--k8s_url` | Kubernetes API server | (none) |
| `K8S_TOKEN` | `--k8s_token` | Kubernetes API token | (none) |


## API

The Amalgam8 Service Registry [API documentation](https://amalgam8.io/registry) is available in Swagger format.

## Contributing

Contributions and feedback are welcome! 
Proposals and Pull Requests will be considered and responded to. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/registry/blob/master/CONTRIBUTING.md)
file for more information.

### Contributing Changes

Go code contributed to Amalgam8 Service Registry must use default Go formatting and pass: 

* [golint](https://github.com/golang/lint)
* Go vet

These actions are run for you by invoking
```sh
make precommit
```

You can install a git-hook into the local `.git/hooks/` directory, as a pre-commit ot pre-push hook.  

## License

Copyright 2016 IBM Corporation
 
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance 
with the License. You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0  

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed 
on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. 
See the License for the specific language governing permissions and limitations under the License.
