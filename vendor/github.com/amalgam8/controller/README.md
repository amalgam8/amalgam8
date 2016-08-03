# Amalgam8 Controller

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/controller
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/controller
[Travis]: https://travis-ci.org/amalgam8/controller
[Travis Widget]: https://travis-ci.org/amalgam8/controller.svg?branch=master

The Amalgam8 Controller serves as a central configuration service in the Amalgam8
platform that allows one to setup rules for routing traffic across
different microservice versions, rules for injecting faults into
microservice API calls, etc., based on various request-level attributes.

In addition to route management, the Amalgam8 Controller automatically synchronizes
service instance information from the
[Amalgam8 Registry](https://github.com/amalgam8/registry) and updates the
[Amalgam8 Sidecars](https://github.com/amalgam8/sidecar) attached to each
microservice.

By default, the Amalgam8 Controller operates without any authentication. It also
supports two authentication mechanisms: a trusted auth mode for local
testing and development, and a JWT auth mode for production deployments.

See https://www.amalgam8.io for an overview of the Amalgam8 project
and https://www.amalgam8.io/docs for detailed documentation.

## Usage

To get started, use the current stable version of the Amalgam8 Controller from Docker
Hub.

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
| A8_API_PORT | --api_port | API port | 6379 |
| A8_CONTROL_TOKEN | --control_token | Controller API authentication token | ABCDEFGHIJKLMNOP |
| A8_ENCRYPTION_KEY | --encryption_key | secret key | abcdefghijklmnop |
| A8_POLL_INTERVAL | --poll_interval | poll interval | 10s |
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

The documentation for
[Amalgam8 Controller's REST API](https://amalgam8.io/controller) is
available in Swagger format.

## Building from source

To build from source, clone this repository, and follow the instructions below.

### Pre-requisites

* Docker engine >=1.10
* Go toolchain (tested with 1.6.x). See [Go downloads](https://golang.org/dl/) and [installation instructions](https://golang.org/doc/install).


### Building a Docker Image

To build the docker image for the Amalgam8 Controller service, run the
following commands:

```bash
cd $GOPATH/src/github.com/amalgam8/controller
make build docker
```

You should now have a docker image tagged `a8-controller:latest`.

### Building an Executable

The Amalgam8 Controller can also be run outside of a docker container as a Go
binary.  This is not recommended for production, but it can be useful for
development or easier integration with your local Go tools.

The following commands will build and run it as a Go binary:

```
cd $GOPATH/src/github.com/amalgam8/controller
make build
./bin/controller
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


### Continuous Integration with Travis CI

Continuous builds are run on Travis CI. These builds use the `.travis.yml` configuration.

## Release Workflow

This section includes instructions for working with releases, and is intended for the project's maintainers (requires write permissions)

### Creating a release

1.  Set a version for the release, by incrementing the current version
    according to the [semantic versioning](https://semver.org/)
    guidelines. For example,

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

1.  Edit the [GitHub release object](https://github.com/amalgam8/controller/releases), and add a title and description (according to `CHANGELOG.md`).

## License
Copyright 2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Contributing

Contributions and feedback are welcome!
Proposals and pull requests will be considered.
Please see the [CONTRIBUTING.md](https://github.com/amalgam8/amalgam8.github.io/blob/master/CONTRIBUTING.md) file for more information.
