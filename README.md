# Amalgam8 Service Discovery Registry

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/registry
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/registry
[Travis]: https://travis-ci.org/amalgam8/registry
[Travis Widget]: https://travis-ci.org/amalgam8/registry.svg?branch=master

The Amalgam8 Service Discovery Registry is software developed for registering and locating instances in applications 
built using a microservice based architecture. The registry can be used to provide multi-tenant registration and 
discovery namespaces (i.e., isolation between different tenant scopes), a unified discovery for service
instances registered using different backends (e.g., Kubernetes, Eureka), and different authentication and authorization
backends. 

Note: Amalgam8 service registry is currently in alpha. This means the API will change, features will be added/removed 
and there will be bugs. 

## Building and Running

A recent prebuilt Amalgam8 registry image can be found in docker.io/amalgam8/a8-registry:0.1.
```sh
$ docker pull amalgam8/a8-registry:0.1
```

If you wish to build the registry from source, clone this repository and follow 

### Preprequisites
A developer tested development environment requires the following:
* Linux host (tested with Ubuntu 16.04 LTS)
* Docker (tested with 1.10)
* Go toolchain (tested with 1.6.x). See [Golang downloads](https://golang.org/dl/) and [installation instructions](https://golang.org/doc/install).


### Building a Docker Image

The Amalgam8 registry may be built by simply typing `make build docker` with the [Docker
daemon](https://docs.docker.com/installation/) (v1.10.x) running.

This will produce an image tagged `registry:0.1` which you may run as described below.

### Standalone

The Amalgam8 registry may also be run outside of a docker container as a Go binary. 
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

The registry supports a number of configuration options, most of which should be set through environment variables.

The environment variables can be set via command line flags as well 

### Command Line Flags and Environment Variables

The following environment variables are available. All of them are optional.

#### Registry Configuration

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `API_PORT` | `--api_port` | 80 | API port number | 8080 |
| `LOG_LEVEL` | `--log_level` | `info` | Logging level. Supported values are: `debug`, `info`, `warn`, `error`, `fatal`, `panic` | `debug` |
| `LOG_FORMAT` | `--log_format` | `json` | Logging format. Supported values are: `text`, `json`, `logstash` | `text` |
| `NAMESPACE_CAPACITY` | `--namespace_capacity` | 100 | maximum number of instances that may be registered in a namespace | -1 (no capacity limit) |  
| `DEFAULT_TTL` | `--default_ttl` | 1m | Registry default instance time-to-live (TTL) | 30s |
| `MIN_TTL` | `--min_ttl` | 10s | Minimum TTL that may be specified during registration | 10s | 
| `MAX_TTL` | `--max_ttl` | 20m | Maximum TTL that may be specified during registration | 10m |


#### Authentication and Authorization

The Amalgam8 Service Registry optionally supports multi-tenancy, by isolating each tenant into a separate namespace.
A namespace is defined by an opaque string carried in the HTTP `Authorization` header of API requests. The following
namespace authorization methods are supported and controlled via the `AUTH_MODE` environment variable (or `--auth_mode`
flag):
* None: if no authorization mode is defined, all instances are registered into a default shared namespace. 
* Trusted: namespace is retrieved directly from the Authorization header. This provides namespace separation in a trusted
environment (e.g., single tenant with multiple applications or environments)
* JWT: encodes the namespace value in a signed JWT token claim. 

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `AUTH_MODE` | `--auth_mode` | `jwt` | Authentication modes. Supported values are: `trusted`, `jwt` | none (no isolation) |
| `JWT_SECRET` | `--jwt_secret` | `53cr3t` | Secret key for JWT authentication | none (must be set if `AUTH_MODE` is `jwt`) |
| `REQUIRE_HTTPS` | `--require_https` | `true` | Require clients to use HTTPS for API calls | `false` |

If `jwt` is specified, `JWT_SECRET` (or `--jwt_secret`) must be set as well to allow encryption and decryption.
Namespace value encoding must be present in every API call using HTTP Bearer Authorization:

```sh
Authorization: Bearer jwt.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ...wifQ.Gbz4G_O...NqdY`
```

#### Clustering
Amalgam8 Service Discovery uses a memory only storage solution, without persistency (although different storage 
backends can be implemented). To provide HA and scale, the registry can be run in a cluster and supports replication
between cluster members.

Peer discovery currently uses a shared volume between all members. The volume must be mounted RW into each container.
Alternative discovery mechanisms are being explored.

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `CLUSTER_SIZE` | `--cluster_size` | 3 | Cluster minimal healthy size, peers detecting a lower value will log errors | 1 (standalone) |
| `CLUSTER_DIR` | `--cluster_dir` | /tmp/sd | Filesystem directory for cluster membership | |
| `REPLICATION` | `--replication` | `true` | Enable replication between cluster members | `false` |
| `REPLICATION_PORT` | `--replication_port` | 8081 | Replication port number | 6100 |
| `SYNC_TIMEOUT` | `--sync_timeout` | 60s | Timeout for establishing connections to peers for replication | 30s |


## API

Service Discovery [API documentation](https://amalgam8.io/registry) is available in Swagger format.

## Contributing

Contributions and feedback are welcome! 
Proposals and Pull Requests will be considered and responded to. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/registry/blob/master/CONTRIBUTING.md)
file for more information.

### Contributing Changes

Go code contributed to Amalgam8 Service Discovery Registry must use default Golang formatting and pass: 

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

