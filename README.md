# sidecar

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/sidecar
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/sidecar
[Travis]: https://travis-ci.org/amalgam8/sidecar
[Travis Widget]: https://travis-ci.org/amalgam8/sidecar.svg?branch=master

A language agnostic sidecar for building microservice applications with
automatic service registration, and load-balancing

### Architecture

![Sidecar architecture](https://github.com/amalgam8/sidecar/blob/master/sidecar.jpg)

## Usage
A prebuild Docker iamge is available. Install Docker 1.8 or 1.9 and run the following:

```docker pull amalgam8/a8-controller```

### Configuration options
Configuration options can be set through environment variables or command line flags. 

| Environment Key | Flag Name                   | Description | Default Value |
|:----------------|:----------------------------|:------------|:--------------|
| LOG_LEVEL | --log_level | Logging level (debug, info, warn, error, fatal, panic) | info |
| SERVICE | --service | service name to register with | |
| SERVICE_VERSION | --service_version | service version to register with |  |
| ENDPOINT_HOST | --endpoint_host | service endpoint host name |  |
| ENDPOINT_PORT | --endpoint_port | service endpoint port | |
| REGISTER | --register | enable automatic service registration and heartbeat |  |
| PROXY | --proxy | enable automatic service discovery and load balancing across services using NGINX |  |
| LOG | --log | enable logging of outgoing requests through proxy using FileBeat |  |
| SUPERVISE | --supervise | Enable monitoring of application process. If application dies, container is killed as well. This has to be the last flag. All arguments provided after this flag will considered as part of the application invocation |  |
| TENANT_ID | --tenant_id | service Proxy instance GUID |  |
| TENANT_TOKEN | --tenant_token | token for Service Proxy instance |  |
| TENANT_TTL | --tenant_ttl | tenant TTL for Registry | 1m0s |
| TENANT_HEARTBEAT | --tenant_heartbeat | tenant heartbeat interval to Registry |  |
| REGISTRY_URL | --registry_url | registry URL | 45s |
| REGISTRY_TOKEN | --registry_token | registry API token | |
| NGINX_PORT | --nginx_port | port for NGINX | 6379 |
| CONTROLLER_URL | --controller_url | controller URL |  |
| CONTROLLER_POLL | --controller_poll | interval for polling Controller | 15s |
| LOGSTASH_SERVER | --logstash_server | logstash target for nginx logs |  |
| KAFKA_USER | --kafka_user | kafka username |  |
| KAFKA_PASS | --kafka_pass | kafka password |  |
| KAFKA_TOKEN | --kafka_token | kafka token |  |
| KAFKA_ADMIN_URL | --kafka_admin_url | kafka admin URL |  |
| KAFKA_REST_URL | --kafka_rest_url | kafka REST URL |  |
| KAFKA_SASL | --kafka_sasl | use SASL/PLAIN authentication for kafka |  |
| KAFKA_BROKER | --kafka_broker [--kafka_broker option --kafka_broker option] | kafka brokers |  |
|  | --help, -h | show help | |
|  | --version, -v | print the version | |

## Build from source
The follow section describes options for building the sidecar from source. Instructions on using a prebuilt Docker image are available [here](https://github.com/amalgam8/sidecar#usage).

### Preprequisites
* Docker 1.8 or 1.9
* Go 1.6

### Clone

Clone the repository manually, or use `go get`:

```go get github.com/amalgam8/sidecar```

### Make targets
The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description |
|:-----------------|:------------|
| `release`        | *(Default)* `release` builds the sidecar within a docker container and packages it into a image |
| `test`           | `test` runs all tests using `go test` |
| `clean`          | `clean` removes build artifacts. *Note: this does not remove docker images* |

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
