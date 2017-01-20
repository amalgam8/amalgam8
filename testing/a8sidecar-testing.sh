#!/bin/bash
#
# Copyright 2017 IBM Corporation
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
#
# Amalgam8 Sidecar installation script for Ubuntu/Debian/Centos/Fedora/RHEL distributions.

set -x
set -e

##Install OpenResty from Amalgam8 repo
## Compared to OpenResty stock configuration, this binary has been compiled to place config files in /etc/nginx,
## log files in /var/log/nginx and nginx binary in /usr/sbin/nginx.

A8TMP="/tmp/a8tmp"
mkdir -p $A8TMP

tar -xzf /opt/microservices/a8sidecar-current.tar.gz -C $A8TMP
tar -xzf $A8TMP/opt/openresty_dist/*.tar.gz -C /

#Install Sidecar -- This should be in the end, as it overwrites default nginx.conf, filebeat.yml
tar -xzf /opt/microservices/a8sidecar-current.tar.gz -C /

#Cleanup
rm -rf ${A8TMP}
rm /opt/microservices/a8sidecar-current.tar.gz
