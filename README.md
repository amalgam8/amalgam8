# sidecar

A language agnostic sidecar for building microservice applications with
automatic service registration, and load-balancing

### Architecture

![Sidecar architecture](https://github.com/amalgam8/sidecar/blob/master/sidecar.jpg)

### Environment variables needed to run sidecar
    
* ENDPOINT_HOST, ENDPOINT_PORT -- IP and port of service instance to register
* SERVICE -- Name of service to register
* SD_URL, SD_TOKEN -- URL and auth token for use with service discovery
* RE_URL -- Service proxy control plane URL
* SP_TENANT_ID, SP_TENANT_TOKEN - ID and auth token for use with service  proxy
  
#### IBM MessageHub integration - environment variables
* VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_KAFKA_REST_URL
* VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_API_KEY
* VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_KAFKA_BROKERS_SASL_[0,1,2,3..]
* VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_USER
* VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_PASSWORD

### Running sidecar

Command line arguments
* -register - enable automatic service registration
* -proxy - enable nginx service proxy
* -log - use filebeat to propagate nginx logs to logstash
* -supervise - invoke and monitor application process

Usage:
```bash
sidecar -register -proxy -log -supervise myapp arg1 arg2 -arg3=3 -arg4=4
```
