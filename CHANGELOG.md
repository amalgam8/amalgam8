# Changelog

## 0.3-rc1 (September 1, 2016)

- The controller API has been overhauled to support a wider range of routing and fault injection rules.

- The dependency on Kafka has been removed from controller and sidecar.

- Sidecar now polls controller for rule changes and registry for service instance changes.

- Controller now has support for Redis persistent storage.

- NGINX Lua code overhauled to support new rules API format from controller.

- Options for sidecar to proxy, register, and log are now default `false`.

- `A8_SERVICE` is now in the form of `<service_name>:<tag1>,<tag2>,...,<tagN>` 
where `<tagN>` can be a version number or any other tag.  Sidecar will register
with registry using these tags and rules can be defined to target services 
with a particular set of tag(s).

- Controller now supports single tenant (default), trusted, and JWT authentication.

- Default authentication behavior of controller and registry is `global auth`,
wherein they are configured to run in single tenant mode without any authentication.
In this scenario, `A8_CONTROLLER_TOKEN` and `A8_REGISTRY_TOKEN` should not be 
provided to the sidecar.

## 0.2.1 (August 1, 2016)

- Fixed [#33](https://github.com/amalgam8/registry/issues/33): Panic during synchronization of the Registry's store 

- Fixed bug where controller did not send rule updates to sidecar in
  polling mode.

## 0.2.0 (July 20, 2016)

- The controller no longer generates Nginx config files in response to
  changes in the service instance pool. Instead, these 
  updates are passed on to the sidecar which in turn passes them onto the
  Lua code that now manages the upstreams.  This decoupling of
  configuration management from the controller provides the user with full
  control over the nginx configuration in each sidecar, allowing her to
  customize the configuration (e.g., add HTTPS certificates, custom
  modules, etc.).

- All environment variables pertaining to the controller and sidecar are now prefixed
  with A8_.

- TENANT_ID is no longer needed nor accepted. Tenants are distinguished by
  unique authentication tokens that needs to be present in each request to
  the controller.

- Nginx proxy is no longer reloaded whenever there are changes in
  microservices/instances. Using `balance_by_lua`, Nginx configuration is
  dynamically updated.

- Support for registry adapters that can be used to synchronize the
  registry state with other service registries. The K8S adapter
  automatically synchronizes service registration information stored in
  etcd.

- Support for statically registering TCP microservices specified in a
  configuration file.

- Support for registering HTTPS endpoints

- Several bug fixes


## 0.1.0 (June 28, 2016)
- Initial release of the Amalgam8