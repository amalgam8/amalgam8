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

# This is a temporary hack to create openresty debian packages until a 
# maintainable solution is in place, i.e., a hosted deb package.

set -x
set -e
## These RPM packages are built by the OpenResty team and hosted at https://openresty.org/en/linux-packages.html
## Based on the description in https://openresty.org/en/rpm-packages.html, the key packages needed are openresty-1.11.2.1, openresty-openssl-1.0.2h, 
## openresty-pcre-8.39 and openresty-zlib-1.2.8. The last three packages may be available debian distributions but they might be outdated.

## WARNING: THESE URLS MAY CHANGE ANYTIME. Check https://copr-be.cloud.fedoraproject.org/results/openresty/openresty/fedora-24-x86_64/ for up-to-date packages
wget https://copr-be.cloud.fedoraproject.org/results/openresty/openresty/fedora-24-x86_64/00446691-openresty/openresty-1.11.2.1-3.fc24.x86_64.rpm
wget https://copr-be.cloud.fedoraproject.org/results/openresty/openresty/fedora-24-x86_64/00444092-openresty-openssl/openresty-openssl-1.0.2h-5.fc24.x86_64.rpm
wget https://copr-be.cloud.fedoraproject.org/results/openresty/openresty/fedora-24-x86_64/00444013-openresty-pcre/openresty-pcre-8.39-2.fc24.x86_64.rpm
wget https://copr-be.cloud.fedoraproject.org/results/openresty/openresty/fedora-24-x86_64/00444012-openresty-zlib/openresty-zlib-1.2.8-1.fc24.x86_64.rpm

## Convert the RPMs to Deb using fpm
fpm -s rpm -t deb openresty-1.11.2.1-3.fc24.x86_64.rpm
fpm -s rpm -t deb openresty-pcre-8.39-2.fc24.x86_64.rpm
fpm -s rpm -t deb openresty-zlib-1.2.8-1.fc24.x86_64.rpm
fpm -s rpm -t deb openresty-openssl-1.0.2h-5.fc24.x86_64.rpm

## The resulting packages are
# openresty_1.11.2.1-3.fc24.x86_64.deb
# openresty-pcre_8.39-2.fc24.x86_64.deb
# openresty-zlib_1.2.8-1.fc24.x86_64.deb
# openresty-openssl_1.0.2h-5.fc24.x86_64.deb
