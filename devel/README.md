# Amalgam8 Developer Guide

## Requirements

* [Go](http://golang.org/) version 1.7.1
* [Docker](https://docs.docker.com/engine/installation/) 1.10 or later
* [Docker Compose](https://docs.docker.com/compose/install/)  1.5.1 or later
* Python 2.7 with pip

The `Vagrantfile` in this folder can be used to instantiate an Ubuntu VM
that has all the dependencies (except Kubernetes). Just do a `vagrant up`
in this folder, followed by `vagrant ssh` to get started.

If you plan on using [Kubernetes](https://kubernetes.io) for managing your
containers, then a recent installation of kubernetes (v1.2.3 or higher) is
needed. The accompanying helper script `devel/install-kubernetes.sh` can be
used to install Kubernetes inside the vagrant environment.

Once the prerequisites are installed or the Vagrant VM is up, source code
related to Amalgam8 sidecar, controller and registry can be found under the
`$GOPATH/src/github.com/amalgam8/amalgam8` folder. The `a8ctl` tool can be
found under the `$GOPATH/src/github.com/amalgam8/a8ctl` folder.

## Repository Structure

* The source code is organized into Go packages for the sidecar, service
registry and the route controller.

* The examples folder contains the source code for the demo applications
(helloworld and bookinfo).

* The testing folder contains test scripts used for integration testing in
Travis builds.

## Building Docker Images

To build the docker image for the Amalgam8 Route Controller, Service
Registry and the Amalgam8 Sidecar, run the following commands:

```bash
cd $GOPATH/src/github.com/amalgam8/amalgam8
make build dockerize
```

You should now have three docker images, namely `a8-controller:latest`,
`a8-registry:latest` and `a8-sidecar:latest`.

### Demo Applications

From the `$GOPATH/src/github.com/amalgam8/amalgam8` folder, to build the
docker images for the [demo applications](https://github.com/amalgam8/amalgam8/blob/master/examples/),
run the following scripts:

```bash
#for the helloworld app
./examples/apps/helloworld/build-services.sh
#for the bookinfo app
./examples/apps/bookinfo/build-services.sh
```

## Building Executables

The Go-based components can also be run outside of a docker container as Go
binaries.  While this is not recommended for production, it can be useful
for development or easier integration with your local Go tools.

The `make build` command builds binaries for the controller, registry and
the sidecar. The binaries can be found under the `bin` folder under the
repository root directory (`$GOPATH/src/github.com/amalgam8/amalgam8`).


## Makefile Targets

The following Makefile targets are available for the `amalgam8` repository.

| Make Target      | Description |
|:-----------------|:------------|
| `build`          | *(Default)* invokes targets `build.controller`, `build.registry`, `build.sidecar` |
| `build.controller`        | builds the controller binary in the ./bin directory |
| `build.registry`          | builds the registry binary in the ./bin directory |
| `build.sidecar`          |  builds the sidecar binary in the ./bin directory |
| `dockerize`         | invokes targets `dockerize.controller`, `dockerize.registry`, `dockerize.sidecar` |
| `dockerize.controller`         | packages the controller binary into a docker container with tag `a8-controller:latest` |
| `dockerize.registry`         | packages the registry binary into a docker container with tag `a8-registry:latest` |
| `dockerize.sidecar`         | packages the sidecar binary into a docker container with tag `a8-sidecar:latest` |
| `release`        | invokes targets `release.controller`, `release.registry`, `release.sidecar` |
| `release.controller`        | creates a tarball with the controller binary |
| `release.registry`        | creates a tarball with the registry binary |
| `release.sidecar`        | creates a tarball with the sidecar binary, openresty binaries, and nginx configuration files required to run the sidecar|
| `precommit`      | runs code formatting and checks. `make precommit` must be run before committing code changes and sending pull requests |
| `test`           | runs (short duration) tests using `go test`. You can also use `make test.long` to include long running tests. |
| `test.integration`           | runs end to end integration tests using docker-compose and kubernetes. The test scripts can be found under the testing folder relative to root of repository. |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |


## Building the CLI

To use the latest and potentially unstable version of the `a8ctl` CLI, run
the following command from `$GOPATH/src/github.com/amalgam8/a8ctl` folder:

```bash
python setup.py develop
```

---

### Cleanup

To remove locally-compiled Amalgam8 images and use the Amalgam8 images from Docker Hub:

```bash
docker rmi $(docker images | grep "amalgam8/" | awk "{print \$3}")
```
