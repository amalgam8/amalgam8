# Changelog

## 0.3.2 (October 12, 2016)

- Sidecars now support HTTP health checks. The health check endpoint can be
  provided as part of the sidecar configuration file. When the application
  fails to respond to a health check, it is unregistered from the service
  registry, causing the instance to be removed from the load balancing pool
  of other sidecars upon the next refresh.

- A special CLI mode for the `a8sidecar` binary that allows users to view
  the state inside the sidecar. ([PR#335](https://github.com/amalgam8/amalgam8/pull/335))

- Minor performance optimizations to the Lua code in the sidecar and bug
  fixes. Timeout support is currently disabled. Specifying timeouts in the
  rules will not have any effect. ([PR#334](https://github.com/amalgam8/amalgam8/pull/334))

- Registry now supports Eureka remote catalog, similar to Kubernetes
  catalog. Service registration info in Eureka will now be synced with the
  Amalgam8 service registry automatically. ([PR#247](https://github.com/amalgam8/amalgam8/pull/247))

- Eureka metadata tags are automatically translated into Amalgam8
  instance tags in the service registry.

- Optimizations to the redis operations used in registry code

- The bookinfo example application is now a polyglot application, composed
  of services written in Java, Ruby, and Python.

- Simplification of the demo scripts: consolidate into fewer files and eliminate
  unnecessary scripts.

- All documentation has been moved to https://amalgam8.io/docs/

## 0.3.1 (September 20, 2016)

- Amalgam8 nginx configuration files in the sidecar is now split into
multiple files. Amalgam8 specific code has been abstracted away into
separate files, and user-customizable part is now confined to location
blocks in amalgam8-services.conf ([PR#278](https://github.com/amalgam8/amalgam8/pull/278))

- Fixed invalid DNS config in kubernetes config that caused code compiled
with Go 1.7.1 to fail ([PR#280](https://github.com/amalgam8/amalgam8/pull/280)).

- Fixed bug in sidecar that caused HTTP 500 when version cookie did not
match any backend in the route list ([PR#271](https://github.com/amalgam8/amalgam8/pull/271)).

- Fixed bugs in bluemix deployment scripts and updated READMEs to point to
the correct version of Bluemix CLI
([PR#275](https://github.com/amalgam8/amalgam8/pull/275) and
[PR#279](https://github.com/amalgam8/amalgam8/pull/279)).

## 0.3.0 (September 12, 2016)

- The controller API has been overhauled to support a wider range of routing and fault injection rules.

- The dependency on Kafka has been removed from controller and sidecar.

- Sidecar now polls controller for rule changes and registry for service instance changes.

- Controller and Registry now have support for Redis persistent storage.

- NGINX Lua code has been overhauled to support the new rules API format from controller.

- Options for sidecar to proxy, register, and log now default to `false`.

- The default controller port is now 8080 instead of 6379.

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

- Fixed [#33](https://github.com/amalgam8/amalgam8/issues/162): Panic during synchronization of the Registry's store 

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
