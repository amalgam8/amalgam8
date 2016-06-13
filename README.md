# Amalgam8 Controller

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/controller
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/controller
[Travis]: https://travis-ci.org/amalgam8/controller
[Travis Widget]: https://travis-ci.org/amalgam8/controller.svg?branch=master

The Amalgam8 controller is software developed for managing load balancing and failure injection in applications built using a microservice based architecture.

An overview of the Amalgam8 project is available here: http://amalgam8.io/

## Usage

A prebuilt Docker image is available. Install Docker 1.8 or 1.9 and run the following:

```docker pull amalgam8/a8-controller```

Provide [configuration options]() as environment variables. For example:

```docker run amalgam8/a8-controller -e poll_interval=60s```

### Configuration options
Configuration options can be set through environment variables or command line flags. 

| Environment Key | Flag Name                   | Example Value(s)            | Description | Default Value |
|:----------------|:----------------------------|:----------------------------|:------------|:--------------|
| `key` | `-f | --flag` | Sample value for this option | Description. | none |
| `etc` | `-e` | anything | Something | undefined |

### REST API

Documentation is available in [Swagger](https://github.com/amalgam8/controller/blob/master/swagger.json) format.

## Build from source
The follow section describes options for building the controller from source. A prebuilt Docker image is available [here](http://xxx).

### Preprequisites
* Docker 1.8 or 1.9
* Go 1.6

### Clone

Clone the repository manually, or use `go get`:

```go get github.com/amalgam8/controller```

### Make targets
The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `release`        | *(Default)* `release` builds the controller within a docker container and packages it into a image |
| `test`           | `test` runs all tests using `go test` |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

### Build Docker image
Docker must be configured and running to build a Docker image.
```
cd $GOPATH/src/github.com/amalgam8/controller
make build
make docker IMAGE=controller
```

This will produce an image tagged `controller:latest`.

### Build executable
The Docker image is the recommended way of deploying the controller. However, manually building the executable may be useful for testing.

The following commands will build and run the controller:

```target
make build
./bin/controller
```

## License
Copyright 2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Contributing

Contributions and feedback are welcome! 
Proposals and pull requests will be considered. Please see the
[CONTRIBUTING.md](https://github.com/amalgam8/controller/blob/master/CONTRIBUTING.md)
file for more information.
