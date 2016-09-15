# Amalgam8 Registry

The Amalgam8 Registry is a multi-tenant, highly-available service for
service registration and service discovery in microservice applications.

**High Availability:** For high availability, run the Registry in clustered
mode, where data in memory is replicated across registry
instances. Alternatively, the Registry supports a Redis backend for storing
instance information. In this mode, the Registry instances are stateless
and can be scaled using standard autoscaling techniques. In tandem, the
data store backend, i.e., Redis, needs to be deployed in HA mode. Please
refer to the [Redis Clustering](http://redis.io/topics/cluster-tutorial)
documentation for details on setting up a highly available Redis backend.

In both cases, the Registry provides eventual consistency for data
synchronization across Registry instances.

**Extensibility:** The Registry is built using an extensible catalog model
that allows the user to update the Amalgam8 Registry with information
stored in other service registries such as Kubernetes, Consul, etc. In
addition, the Registry can be used as a drop-in replacement for
[Netflix Eureka](https://github.com/Netflix/eureka) with full API
compatibility. Eureka compatibility enables the Registry to be used with
Java clients using [Netflix Ribbon](https://github.com/Netflix/ribbon).

**Authentication & Multi-tenancy:** By default, the Registry operates in a
single-tenant mode without any authentication. In multi-tenant mode, it
supports two authentication mechanisms: a trusted auth mode for local
testing and development purposes, and a JWT auth mode for production
deployments.

See https://www.amalgam8.io/docs for detailed documentation.


## Usage

To get started, use the current stable version of the Registry from Docker
Hub.

```bash
docker run amalgam8/a8-registry:latest
```

### Command Line Flags and Environment Variables

The Amalgam8 Registry supports a number of configuration options,
most of which are set through environment variables. The environment
variables can be set via command line flags as well.

The following environment variables are available. All of them are optional.

#### Registry Configuration

| Environment Variable | Flag Name                   | Description | Default Value |
|:---------------------|:----------------------------|:------------|:--------------|
| `A8_API_PORT` | `--api_port` | API port number | 8080 |
| `A8_LOG_LEVEL` | `--log_level` | Logging level. Supported values are: `debug`, `info`, `warn`, `error`, `fatal`, `panic` | `debug` |
| `A8_LOG_FORMAT` | `--log_format` | Logging format. Supported values are: `text`, `json`, `logstash` | `text` |
| `A8_NAMESPACE_CAPACITY` | `--namespace_capacity` | maximum number of instances that may be registered in a namespace | -1 (no capacity limit) |  
| `A8_DEFAULT_TTL` | `--default_ttl` | Registry default instance time-to-live (TTL) | 30s |
| `A8_MIN_TTL` | `--min_ttl` | Minimum TTL that may be specified during registration | 10s | 
| `A8_MAX_TTL` | `--max_ttl` | Maximum TTL that may be specified during registration | 10m |


#### Authentication and Authorization

The Amalgam8 Registry supports multi-tenancy by isolating each tenant into a separate namespace.
A namespace is defined by an opaque string carried in the HTTP `Authorization` header of API requests. The following
namespace authorization methods are supported and controlled via the `A8_AUTH_MODE` environment variable (or `--auth_mode`
flag):
* None: if no authorization mode is defined, all instances are registered into a default shared namespace. 
* Trusted: namespace is retrieved directly from the Authorization header. This provides namespace separation in a trusted
environment (e.g., single tenant with multiple applications or environments).
* JWT: encodes the namespace value in a signed JWT token claim. 

| Environment Variable | Flag Name                   | Description | Default Value |
|:---------------------|:----------------------------|:------------|:--------------|
| `A8_AUTH_MODE` | `--auth_mode` | Authentication modes. Supported values are: `trusted`, `jwt` | none (no isolation) |
| `A8_JWT_SECRET` | `--jwt_secret` | Secret key for JWT authentication | none (must be set if `A8_AUTH_MODE` is `jwt`) |
| `A8_REQUIRE_HTTPS` | `--require_https` | Require clients to use HTTPS for API calls | `false` |

If `jwt` is specified, `A8_JWT_SECRET` (or `--jwt_secret`) must be set as well to allow encryption and decryption.
Namespace value encoding must be present in every API call using HTTP Bearer Authorization:

```bash
Authorization: Bearer jwt.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ...wifQ.Gbz4G_O...NqdY`
```

#### Clustering

Amalgam8 Registry uses a memory only storage solution, without persistency (although different storage 
backends can be implemented). To provide HA and scale, the Registry can be run in a cluster and supports replication
between cluster members. To use a persistent storage backend, see the section `Persistent Backend Storage`.

Peer discovery currently uses a shared volume between all members. The
volume must be mounted RW into each container.  We are exploring
alternative discovery mechanisms.


| Environment Variable | Flag Name                   | Description | Default Value |
|:---------------------|:----------------------------|:------------|:--------------|
| `A8_CLUSTER_SIZE` | `--cluster_size` | Cluster minimal healthy size, peers detecting a lower value will log errors | 1 (standalone) |
| `A8_CLUSTER_DIR` | `--cluster_dir` | Filesystem directory for cluster membership | none, must be specified for clustering to work |
| `A8_REPLICATION` | `--replication` | Enable replication between cluster members | `false` |
| `A8_REPLICATION_PORT` | `--replication_port` | Replication port number | 6100 |
| `A8_SYNC_TIMEOUT` | `--sync_timeout` | Timeout for establishing connections to peers for replication | 30s |

#### Persistent Backend Storage

Amalgam8 Registry supports a Redis backend for storing instance information as an alternative to the memory only clustering and replication option.

| Environment Variable | Flag Name                   | Description | Default Value |
|:---------------------|:----------------------------|:------------|:--------------|
| `A8_STORE` | `--store` | Backing store to use to persist Registry instance information. Supported values are: `redis`, `inmem`  | inmem (in memory) |
| `A8_STORE_ADDRESS` | `--store_address` | Address of the Redis server | none, must be specified to use a Redis store |
| `A8_STORE_PASSWORD` | `--store_password` | Password for the Redis backend | none, assumes no password set for the Redis server |

#### Catalog Extensions

The Amalgam8 Registry supports read-only catalogs extensions. 
The content of each catalog extension (e.g., Kubernetes, Docker-Swarm, Eureka, FileSystem, etc) is read by the Registry and
returned to the user along with the content of the Registry itself.

| Environment Variable | Flag Name                   | Description | Default Value |
|:---------------------|:----------------------------|:------------|:--------------|
| `A8_K8S_URL` | `--k8s_url` | Enable kubernetes catalog and specify the API server | (none) |
| `A8_K8S_TOKEN` | `--k8s_token` | Kubernetes API token | (none) |
| `A8_EUREKA_URL` | `--eureka_url` | Enable eureka catalog and specify the API server(s) | (none) |
| `A8_FS_CATALOG` | `--fs_catalog` | Enable FileSystem catalog and specify the directory of the config files. The format of the file names in the directory should be `<namespace>.conf`. See [FileSystem catalog documentation](doc/filesystem_catalog.md) for more information | (none) |


## REST API

Amalgam8 Registry [API documentation](../api/swagger-spec/registry.json) is
available in Swagger format.

## Building from source

Please refer to the [developer guide](../devel/) for prerequisites and
instructions on how to setup the development environment.

### Building a Docker Image

To build the docker image for the Amalgam8 Registry service, run the
following commands:

```bash
cd $GOPATH/src/github.com/amalgam8/amalgam8
make build dockerize.registry
```

You should now have a docker image tagged `a8-registry:latest`.

### Building an Executable

The Amalgam8 Registry can also be run outside of a docker container as a Go
binary.  This is not recommended for production, but it can be useful for
development or easier integration with your local Go tools.

The following commands will build and run it as a Go binary:

```bash
cd $GOPATH/src/github.com/amalgam8/amalgam8
make build.registry
./bin/a8registry
```

### Makefile Targets

The following Makefile targets are available.

| Make Target      | Description |
|:-----------------|:------------|
| `build`          | *(Default)* `build` builds the Registry binary in the ./bin directory |
| `precommit`      | `precommit` should be run by developers before committing code changes. It runs code formatting and checks. |
| `test`           | `test` runs (short duration) tests using `go test`. You may also `make test.all` to include long running tests. |
| `docker`         | `docker` packages the binary in a docker container |
| `release`        | `release` builds a tarball with the Registry binary |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |
