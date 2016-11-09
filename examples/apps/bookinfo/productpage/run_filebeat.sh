#!/bin/sh
LOGSTASH=`echo $A8_LOGSTASH_SERVER`

if [ -z "$LOGSTASH" ]; then
  echo "Env var A8_LOGSTASH_SERVER is not set."
  exit 1
fi

sed -e s/LOGSTASH_REPLACEME/\"$LOGSTASH\"/ /etc/filebeat/filebeat.yml >/tmp/filebeat.yml
cp /tmp/filebeat.yml /etc/
exec filebeat -c /etc/filebeat.yml