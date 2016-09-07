#!/bin/bash
#
# Copyright 2016 IBM Corporation
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


set -x

if [ "$1" == "" -o "$2" == "" ]; then
    echo "usage: $0 [gateway|bookinfo|helloworld] start|stop"
    exit 1
fi

COMPONENT=$1
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
#MYIP=`ip addr show eth0 | awk '$1 == "inet" {gsub(/\/.*$/, "", $2); print $2}'`
MYIP=192.168.33.33
cp $SCRIPTDIR/${COMPONENT}.json /tmp/
sed -i "s/__REPLACEME__/${MYIP}/" /tmp/${COMPONENT}.json

APP="/tmp/${COMPONENT}.json"
TYPE=groups

if [ "$COMPONENT" == "gateway" ]; then
    TYPE=apps
fi

if [ "$2" == "start" ]; then
    echo "starting ${COMPONENT}"
    cat $APP|curl -X POST -H "Content-Type: application/json" http://${MYIP}:8080/v2/${TYPE} -d@-
elif [ "$2" == "stop" ]; then
    echo "Stopping ${COMPONENT}"
    curl -X DELETE -H "Content-Type: application/json" http://${MYIP}:8080/v2/${TYPE}/${COMPONENT}
else
    echo "usage: $0 start|stop"
    exit 1
fi
