# Changelog

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
