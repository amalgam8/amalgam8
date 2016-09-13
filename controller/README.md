# Amalgam8 Route Controller

The Amalgam8 Route Controller is a multi-tenant, highly-available service
for managing routing across microservices via the
[sidecars](../sidecar/). Routes programmed at the controller are percolated
down to the sidecars periodically. Routing rules can be based on the
content of the requests and the version of microservices sending and
receiving the requests. In addition to routing, rules can also be expressed
for injecting faults into microservice API calls.

**High Availability:** The Route Controller supports a Redis backend for
storing routing information. In this mode, the Controller instances are
stateless and can be scaled using standard autoscaling techniques. In
tandem, the data store backend, i.e., Redis, needs to be deployed in HA
mode. Please refer to the
[Redis Clustering](http://redis.io/topics/cluster-tutorial) documentation
for details on setting up a highly available Redis backend.

**Authentication & Multi-tenancy:** By default, the Controller operates in
a single-tenant mode without any authentication. In multi-tenant mode, it
supports two authentication mechanisms: a trusted auth mode for local
testing and development purposes, and a JWT auth mode for production
deployments.

See https://www.amalgam8.io/docs for detailed documentation.

## Usage

To get started, use the current stable version of the Amalgam8 Controller
from Docker Hub.

```bash
docker run amalgam8/a8-controller:latest
```

### Command Line Flags and Environment Variables

The Amalgam8 Controller supports a number of configuration options, most of
which are set through environment variables. The environment variables can
be set via command line flags as well.

The following environment variables are available. All of them are optional.

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| A8_API_PORT | --api_port | API port | 8080 |
| A8_CONTROL_TOKEN | --control_token | Controller API authentication token | ABCDEFGHIJKLMNOP |
| A8_ENCRYPTION_KEY | --encryption_key | secret key | abcdefghijklmnop |
| A8_DATABASE_TYPE |  --database_type |	database type | memory |
| A8_DATABASE_USERNAME | --database_username | database username | |
| A8_DATABASE_PASSWORD | --database_password | database password | |
| A8_DATABASE_HOST | --database_host | database host | |
| A8_LOG_LEVEL | --log_level | logging level (debug, info, warn, error, fatal, panic) | info |
| A8_AUTH_MODE | --auth_mode | Authentication modes. Supported values are: 'trusted', 'jwt'" | |
| A8_JWT_SECRET | --jwt_secret | Secret key for JWT authentication | |
| | --help, -h | show help | |
| | --version, -v | print the version | |


## REST API

Amalgam8 Route Controller
[API documentation](../api/swagger-spec/controller.json) is available in
Swagger format.

## Building from source

Please refer to the [developer guide](../devel/) for prerequisites and
instructions on how to setup the development environment.

### Building a Docker Image

To build the docker image for the Amalgam8 Route Controller service, run the
following commands:

```bash
cd $GOPATH/src/github.com/amalgam8/amalgam8
make build dockerize.controller
```

You should now have a docker image tagged `a8-controller:latest`.

### Building an Executable

The Amalgam8 Route Controller can also be run outside of a docker container
as a Go binary.  This is not recommended for production, but it can be
useful for development or easier integration with your local Go tools.

The following commands will build and run it as a Go binary:

```bash
cd $GOPATH/src/github.com/amalgam8/amalgam8
make build.controller
./bin/a8controller
```

### Makefile Targets

The following Makefile targets are available.

| Make Target      | Description |
|:-----------------|:------------|
| `build`          | *(Default)* `build` builds the Controller binary in the ./bin directory |
| `precommit`      | `precommit` should be run by developers before committing code changes. It runs code formatting and checks. |
| `test`           | `test` runs (short duration) tests using `go test`. You may also `make test.all` to include long running tests. |
| `docker`         | `docker` packages the binary in a docker container |
| `release`        | `release` builds a tarball with the Controller binary |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |
