# FileSystem Catalog

The FileSystem catalog extension allows the registry to pull static service information from a file in addition to supporting standard registration via REST API.
This catalog is useful when you have services that do not run the Amalgam8 protocol (e.g., database service), but you want other services to discover these services using the standard Amalgam8 discovery REST API.

The name of the config file should be `<namespace>.conf` and the format of the file is a JSON object that contains array of instances.
The format of each instance is the same as the registration body defined in the [API Swagger Documentation](https://amalgam8.io/registry).

To enable this catalog type you have to provide the `--fs_catalog` command line flag and set it to the directory of the config files (one file for each namespace).

Note: If you run multiple registry containers (e.g., for high-availability), you (probably) want to put the config files in a shared volume and to configure that volume for all the registry containers.

## Config File Example

```
{
    "instances": [
        {
            "service_name": "service_a",
            "endpoint": {
                "type" : "tcp",
                "value": "10.0.0.1:8080"
             },
             "status"   : "UP",
             "tags"     : ["v1", "dallas"],
             "metadata" : {
                 "key1" : "value1",
                 "key2" : 90
             }
        },
        {
            "service_name": "service_a",
            "endpoint": {
                "type" : "tcp",
                "value": "10.0.0.2:8080"
             },
             "status"   : "OUT_OF_SERVICE",
             "tags"     : ["v1", "london"],
             "metadata" : {
                 "key1" : "value1",
                 "key2" : 90
             }
        },
        {
            "service_name": "service_b",
            "endpoint": {
                "type" : "tcp",
                "value": "10.1.1.5:8081"
             },
             "status"  : "UP"
        }
    ]
}
```
