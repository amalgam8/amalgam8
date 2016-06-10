# Amalgam8 Controller

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/controller
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/controller
[Travis]: https://travis-ci.org/amalgam8/controller
[Travis Widget]: https://travis-ci.org/amalgam8/controller.svg?branch=master

The Amalgam8 controller is software developed for ... in applications built using a microservice based architecture.

## Image

A recent prebuilt Amalgam8 controller image can be found at TBD.
```sh
$ docker pull amalgam8/controller
```

To build from source instead, see [build](https://github.com/amalgam8/controller#building-from-source).

## Building from source

### Preprequisites

* Docker 1.8 or 1.9
* Go 1.6

### Clone

Clone the repository manually, or use `go get`:

```go get github.com/amalgam8/controller```

### Building a Docker Image

```
cd $GOPATH/src/github.com/amalgam8/controller
make build
make docker IMAGE=controller
```

This will produce an image tagged `controller:latest`.

### Standalone

The Amalgam8 controller may also be run outside of a docker container as a Go binary for testing.

The following commands will build and run the controller:

```
make build
./bin/controller
```

### Make Targets

The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `release`        | *(Default)* `release` builds the controller within a docker container and packages it into a image |
| `test`           | `test` runs all tests using `go test` |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

## Usage

The controller supports a number of configuration options. The configuration options can be set through environment variables or command line flags.

### Command Line Flags and Environment Variables

The following environment variables are available. All of them are optional. They are listed in a general order of likelihood that a user may want to configure them as something other than the defaults.

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `key` | `-f | --flag` | Sample value for this option | Description. | none |
| `etc` | `-e` | anything | Something | undefined |


## REST API

Documentation is available in [Swagger](https://github.com/amalgam8/controller/blob/master/swagger.json) format.

## License
Copyright 2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Contributing

Contributions and feedback are welcome! 
Proposals and Pull Requests will be considered and responded to. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/controller/blob/master/CONTRIBUTING.md)
file for more information.
