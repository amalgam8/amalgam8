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
#
# Amalgam8 Sidecar installation script for Debian distributions.

set -x
set -e

A8SIDECAR_RELEASE=v0.2.0
FILEBEAT_RELEASE=1.2.2
DOWNLOAD_URL=https://github.com/amalgam8/amalgam8/releases/download/${A8SIDECAR_RELEASE}

wget -O /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb https://download.elastic.co/beats/filebeat/filebeat_${FILEBEAT_RELEASE}_amd64.deb
wget -O /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz ${DOWNLOAD_URL}/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz

##Install OpenResty
mkdir /var/log/nginx

mkdir -p /tmp/a8tmp && \
    cd /tmp/a8tmp && \
    tar -xzf /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz && \
    dpkg -i /tmp/a8tmp/openresty_debs/*.deb && \
    apt-get -y update && apt-get -f install \
    ln -s /usr/local/openresty/nginx/sbin/nginx /usr/local/bin/nginx && \
    mkdir -p /etc/nginx && cp /usr/local/openresty/nginx/conf/* /etc/nginx/ && \
    ln -s /usr/local/openresty/nginx/logs /var/log/nginx

#Install Filebeat
dpkg -i /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb

#Install Sidecar -- This should be in the end, as it overwrites default nginx.conf, filebeat.yml
tar -xzf /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz -C /

#Cleanup
rm -rf /tmp/a8tmp
rm /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb
rm /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz
