# Changelog

## 0.2.1 (August 1, 2016)

- Fixed [#33](https://github.com/amalgam8/amalgam8/registry/issues/33): Panic during synchronization of the Registry's store 

## 0.2.0 (July 20, 2016)

- Support for registry adapters that can be used to synchronize the
  registry state with other service registries. The K8S adapter
  automatically synchronizes service registration information stored in
  etcd.

- Support for statically registering TCP microservices specified in a
  configuration file.

- Support for registering HTTPS endpoints

## 0.1.0 (June 28, 2016)
- Initial release of the Amalgam8 Service Registry.
