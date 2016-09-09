# Amalgam8 Developer Guide

## Requirements

* [Go](http://golang.org/) version 1.6
* [Docker](https://docs.docker.com/engine/installation/) 1.10 or later
* [Docker Compose](https://docs.docker.com/compose/install/)  1.5.1 or later
* Python 2.7 with pip

If you plan on using [Kubernetes](https://kubernetes.io) for managing your
containers, then a recent installation of kubernetes (v1.2.3 or higher) is
needed.

The `Vagrantfile` in this folder can be used to instantiate an Ubuntu VM
that has all the dependencies (except Kubernetes). Just do a `vagrant up`
in this folder, followed by `vagrant ssh` to get started.

## Building the Code

Once the prerequisites are installed or the Vagrant VM is up, source code
related to Amalgam8 sidecar, controller and registry can be found under the
`$GOPATH/src/github.com/amalgam8/amalgam8` folder. The `a8ctl` tool can be
found under the `$GOPATH/src/github.com/amalgam8/a8ctl` folder.

### Control Plane

From the `$GOPATH/src/github.com/amalgam8/amalgam8` folder, you can compile
the code and build the docker images for the controller and registry using
the following scripts:

```bash
./testing/build-scripts/build-controller.sh
./testing/build-scripts/build-registry.sh
```

### Sidecar

From the `$GOPATH/src/github.com/amalgam8/amalgam8` folder, to build the
sidecar image, run the following script:

```bash
./testing/build-scripts/build-sidecar.sh
```

### Demo Applications

To build the [demo applications](https://github.com/amalgam8/amalgam8/blob/master/examples/README.md)
use the following script:

```bash
./testing/build-scripts/build-examples.sh
```

### CLI

To use the latest and potentially unstable version of the `a8ctl` CLI, run
the following command from `$GOPATH/src/github.com/amalgam8/a8ctl` folder:

```bash
python setup.py develop
```

---

### Building Everything

To build all three projects and the demo applications, use the following
script from `$GOPATH/src/github.com/amalgam8/amalgam8`

```bash
./testing/build-scripts/build-amalgam8.sh
```

---

### Cleanup

To remove locally-compiled Amalgam8 images and use the Amalgam8 images from Docker Hub:

```bash
docker rmi $(docker images | grep "amalgam8/" | awk "{print \$3}")
```
