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
# Amalgam8 Sidecar installation script for Ubuntu/Debian/Centos/Fedora/RHEL distributions.

set -x
set -e

A8SIDECAR_RELEASE=v0.3.1
DOWNLOAD_URL=https://github.com/amalgam8/amalgam8/releases/download/${A8SIDECAR_RELEASE}
HAVE_WGET=0
HAVE_CURL=0

set +e
curl --version >/dev/null 2>&1
if [ $? -eq 0 ]; then
    HAVE_CURL=1
fi

wget --version >/dev/null 2>&1
if [ $? -eq 0 ]; then
    HAVE_WGET=1
fi

set -e
if [ $HAVE_CURL -eq 0 -a $HAVE_WGET -eq 0 ]; then
    echo "Either curl or wget is required to download the sidecar binary from github"
    exit 1
fi

if [ $HAVE_WGET -eq 1 ]; then
    wget -qO /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz ${DOWNLOAD_URL}/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz
else
    curl -sSL -o /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz ${DOWNLOAD_URL}/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz
fi
  
##Install OpenResty from Amalgam8 repo
## Compared to OpenResty stock configuration, this binary has been compiled to place config files in /etc/nginx,
## log files in /var/log/nginx and nginx binary in /usr/sbin/nginx.

A8TMP="/tmp/a8tmp"
mkdir -p $A8TMP

tar -xzf /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz -C $A8TMP
tar -xzf $A8TMP/opt/openresty_dist/*.tar.gz -C /

#Install Sidecar -- This should be in the end, as it overwrites default nginx.conf, filebeat.yml
tar -xzf /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz -C /

#Cleanup
rm -rf ${A8TMP}
rm /tmp/a8sidecar-${A8SIDECAR_RELEASE}-linux-amd64.tar.gz
