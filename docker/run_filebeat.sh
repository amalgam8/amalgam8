#!/bin/sh
ELASTICSEARCH=`echo $A8_ELASTICSEARCH_SERVER`

if [ -z "$ELASTICSEARCH" ]; then
  echo "Env var A8_ELASTICSEARCH_SERVER is not set."
  exit 1
fi

sed -e s/ELASTICSEARCH_REPLACEME/\"$ELASTICSEARCH\"/ /etc/filebeat/filebeat.yml >/tmp/filebeat.yml
cp /tmp/filebeat.yml /etc/
exec filebeat -e -c /etc/filebeat.yml
