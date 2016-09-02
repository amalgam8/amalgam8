# Changelog

## 0.3-rc1 (September 1, 2016)

- Sidecar now polls controller and registry for changes in rules and 
service instances respectively.

- Removed Kafka dependency from controller and sidecar.

- NGINX Lua code overhauled to support new rules API format from controller.

- Options for sidecar to proxy, register, and log are now default `false`.

- `A8_SERVICE` is now in the form of `<service_name>:<tag1>,<tag2>,...,<tagN>` 
where `<tagN>` can be a version number or any other tag.  Sidecar will register
with registry using these tags and rules can be defined to target services 
with a particular tag.

- Default authentication behavior of controller and registry is `global auth`,
wherein they are configured to run in single tenant mode without any authentication.
In this scenario, `A8_CONTROLLER_TOKEN` and `A8_REGISTRY_TOKEN` should not be 
provided to the sidecar.

## 0.2.0 (July 21, 2016)

- Nginx configuration is no longer managed by the centralized
controller. It is possible to customize the Nginx proxy configuration
(e.g., adding HTTPS certificates, nginx modules, etc.) and
integrate the sidecar into any Docker container.

- Nginx proxy is no longer reloaded whenever there are changes in
microservices/instances. Using `balance_by_lua`, Nginx configuration is
dynamically updated.

- All Amalgam8 related environment variables are now prefixed with A8_ to
keep them separate from other environment variables in the application.

- Several bug fixes

## 0.1.0 (June 28, 2016)
- Initial release of the Amalgam8 Sidecar.
