# Sidecar

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/sidecar
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/sidecar
[Travis]: https://travis-ci.org/amalgam8/sidecar
[Travis Widget]: https://travis-ci.org/amalgam8/sidecar.svg?branch=master

A language agnostic service proxy for building microservice applications with
automatic service registration, and load-balancing.

An overview of the Amalgam8 project is available here: https://amalgam8.io/

Documentation related to the sidecar can be found at
https://amalgam8.io/docs

<!-- ## Architecture -->

<!-- ![Sidecar architecture](https://github.com/amalgam8/sidecar/blob/master/sidecar.jpg) -->

<!-- Refer to the [amalgam8 overview](https://github.com/amalgam8/amalgam8.github.io/blob/master/overview.md#tenant-process) for details. -->

## Usage
A prebuild Docker image is available at Docker Hub. Install Docker 1.8 or 1.9 and run the following:

```docker pull amalgam8/a8-sidecar:0.1```

### Configuration options
Configuration options can be set through environment variables or command
line flags.

**Note:** Atleast one of `-register` or `-proxy` must be enabled.

| Environment Key | Flag Name                   | Description | Default Value |Required|
|:----------------|:----------------------------|:------------|:--------------|--------|
| A8_LOG_LEVEL | --log_level | Logging level (debug, info, warn, error, fatal, panic) | info | no |
| A8_SERVICE | --service | service name to register with | | yes |
| A8_SERVICE_VERSION | --service_version | service version to register with. Service is UNVERSIONED by default |  | needed if you wish to register different versions under same name |
| A8_ENDPOINT_HOST | --endpoint_host | service endpoint hostname. Defaults to the IP (e.g., container) where the sidecar is running | optional |
| A8_ENDPOINT_PORT | --endpoint_port | service endpoint port |  | yes |
| A8_ENDPOINT_TYPE | --endpoint_type | service endpoint type (http, https, udp, tcp, user) | http | no |
| A8_REGISTER | --register | enable automatic service registration and heartbeat | true | See note above |
| A8_PROXY | --proxy | enable automatic service discovery and load balancing across services using NGINX |  | See note above |
| A8_LOG | --log | enable logging of outgoing requests through proxy using FileBeat | true |  | no |
| A8_SUPERVISE | --supervise | Manage application process. If application dies, container is killed as well. This has to be the last flag. All arguments provided after this flag will considered as part of the application invocation | true | no |
| A8_TENANT_TOKEN | --tenant_token | Auth token for Controller instance |  | yes when `-proxy` is enabled |
| A8_TENANT_TTL | --tenant_ttl | TTL for Registry | 60s | no |
| A8_TENANT_HEARTBEAT | --tenant_heartbeat | tenant heartbeat interval to Registry | 45s | no |
| A8_REGISTRY_URL | --registry_url | registry URL |  | yes if `-register` is enabled |
| A8_REGISTRY_TOKEN | --registry_token | registry auth token | | yes if `-register` is enabled |
| A8_NGINX_PORT | --nginx_port | port for NGINX proxy. This port should be exposed in the Docker container. | 6379 | no |
| A8_CONTROLLER_URL | --controller_url | controller URL |  | yes if `-proxy` is enabled |
| A8_CONTROLLER_POLL | --controller_poll | interval for polling Controller | 15s | no |
| A8_LOGSTASH_SERVER | --logstash_server | logstash target for nginx logs |  | yes if `-log` is enabled |
| A8_KAFKA_USER | --kafka_user | kafka username |  | Kafka-based communication with controller is optional |
| A8_KAFKA_PASS | --kafka_pass | kafka password |  | Kafka-based communication with controller is optional |
| A8_KAFKA_TOKEN | --kafka_token | kafka token |  | Kafka-based communication with controller is optional |
| A8_KAFKA_ADMIN_URL | --kafka_admin_url | kafka admin URL |  | Kafka-based communication with controller is optional |
| A8_KAFKA_REST_URL | --kafka_rest_url | kafka REST URL |  | Kafka-based communication with controller is optional |
| A8_KAFKA_SASL | --kafka_sasl | use SASL/PLAIN authentication for kafka |  | Kafka-based communication with controller is optional |
| A8_KAFKA_BROKER | --kafka_broker [--kafka_broker option --kafka_broker option] | kafka brokers |  | Kafka-based communication with controller is optional |
|  | --help, -h | show help | | |
|  | --version, -v | print the version | | |

## Build from source
The following sections describe options for building the sidecar from source. Instructions on using a prebuilt Docker image are available [above](https://github.com/amalgam8/sidecar#usage).

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
| `release`        | *(Default)* `release` builds the sidecar within a docker container and packages it into an image |
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
