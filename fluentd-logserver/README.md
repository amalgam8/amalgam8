* docker build -t logstore_fluentd .
* docker-compose -f compose-fluentd-es.yml up -d
* This will create a fluentd container route its logs to elasticsearch. All docker logs in your system will be routed to fluentd container (kubernetes or not).
* Only logs from docker's stdout and stderr will go to fluentd
* You do not need to start docker with any --log-driver option. This container mounts the /var/lib/docker/containers folder.

This is purely experimental. The preferred approach is to embed filebeat agent in your container and set it up to forward logs to a logstash container,
which in turn can forward logs to elasticsearch. This way, you could run the entire setup locally. For Bluemix, you could change the logstash container to
the one supplied by Opvis (mtlogstash) and it would take care of forwarding logs to elasticsearch (this part is untested).


