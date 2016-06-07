# Amalgam8 Service Discovery Registry

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/registry
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/registry
[Travis]: https://travis-ci.org/amalgam8/registry
[Travis Widget]: https://travis-ci.org/amalgam8/registry.svg?branch=master

The Amalgam8 Service Discovery Registry is software developed for registering and locating instances in applications 
built using a microservice based architecture.

## Building and Running

A recent prebuilt Amalgam8 registry image can be found at TBD.
```sh
$ docker pull amalgam8/registry
```

If you wish to build the registry from source, clone this repository and follow 

### Preprequisites
A developer tested development environment requires the following:
* Linux host (tested with TBD)
* Docker (tested with TBD)
* Go toolchain (tested with TBD). See [TBD](link) for installation instructions.


### Building a Docker Image

The Amalgam8 registry may be built by simply typing `make` with the [Docker
daemon](https://docs.docker.com/installation/) (v1.8.0) running.

This will produce an image tagged `TBD` which you may run as described below.

### Standalone

The Amalgam8 registry may also be run outside of a docker container as a Go binary. 
This is not recommended for production, but it can be useful for development or easier integration with 
your local Go tools.

The following commands will build and run it outside of Docker:

```
make build
./out/registry
```

### Make Targets

The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `release`        | *(Default)* `release` builds the registry within a docker container and and packages it into a scratch-based image |
| `test`           | `test` runs all tests using `go test` |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

## Usage

The registry supports a number of configuration options, most of which should be set through environment variables.

The environment variables can be set via command line flags as well 

### Command Line Flags and Environment Variables

The following environment variables are available. All of them are optional.
They are listed in a general order of likelihood that a user may want to
configure them as something other than the defaults.

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `key` | `-f | --flag` | Sample value for this option | Description. | none |
| `etc` | `-e` | anything | Something | undefined |

## Authentication and Authorization

Explain scope and authentication modes. Refer to command line flags above.
Add example syntax for HTTP

```sh
Authorization: Bearer jwt.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ...wifQ.Gbz4G_O...NqdY`
```

## Clustering

Requirements and set up of clustering (replication). Refer to above?

## API

Service Discovery API documentation is avaialble in Swagger format.

Refer to Swagger in local repo? Link to [Bluemix documentation](https://www.ng.bluemix.net/docs/api/content/api/servicediscovery/rest/index.html)?

## Contributing

Contributions and feedback are welcome! 
Proposals and Pull Requests will be considered and responded to. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/registry/blob/master/CONTRIBUTING.md)
file for more information.

## License

The Amalgam8 Service Discovery Registry is licensed under the TBD License.
