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
set -e

A8SIDECAR_RELEASE=v0.2.0
OPENRESTY_RELEASE=1.9.15.1
FILEBEAT_RELEASE=1.2.2
DOWNLOAD_URL=https://github.com/amalgam8/amalgam8/releases/download/${A8SIDECAR_RELEASE}

apt-get -y update && apt-get -y install libreadline-dev libncurses5-dev libpcre3-dev \
    libssl-dev perl make build-essential curl wget

wget -O /tmp/openresty-${OPENRESTY_RELEASE}.tar.gz https://openresty.org/download/openresty-${OPENRESTY_RELEASE}.tar.gz
wget -O /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb https://download.elastic.co/beats/filebeat/filebeat_${FILEBEAT_RELEASE}_amd64.deb
wget -O /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz ${DOWNLOAD_URL}/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz

##Install OpenResty
adduser --disabled-password --gecos "" nginx
mkdir /var/log/nginx
cd /tmp && \
    tar -xzf /tmp/openresty-${OPENRESTY_RELEASE}.tar.gz && \
    cd /tmp/openresty-${OPENRESTY_RELEASE} && \
    ./configure --with-pcre-jit --with-cc-opt='-O3' --with-luajit-xcflags='-O3' --conf-path=/etc/nginx/nginx.conf --pid-path=/var/run/nginx.pid --user=nginx && \
    make -j2 && \
    make install && \
    ln -s /usr/local/openresty/nginx/sbin/nginx /usr/local/bin/nginx && \
    ldconfig

#Install Filebeat
dpkg -i /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb

#Install Sidecar -- This should be in the end, as it overwrites nginx.conf, filebeat.yml
tar -xzf /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz -C /

#Cleanup
rm -rf /tmp/openresty-${OPENRESTY_RELEASE}
rm /tmp/openresty-${OPENRESTY_RELEASE}.tar.gz
rm /tmp/filebeat_${FILEBEAT_RELEASE}_amd64.deb
rm /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz
