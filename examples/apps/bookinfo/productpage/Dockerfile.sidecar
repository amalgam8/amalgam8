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

FROM python:2-onbuild

ARG A8_SIDECAR_RELEASE

# Install Filebeat
ARG FILEBEAT_VERSION="5.1.1"
RUN set -ex \
    && curl -fsSL https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-${FILEBEAT_VERSION}-linux-x86_64.tar.gz -o /tmp/filebeat.tar.gz \
    && cd /tmp \
    && tar -xzf filebeat.tar.gz \
    \
    && cd filebeat-* \
    && cp filebeat /bin \
    \
    && rm -rf /tmp/filebeat*

COPY filebeat.yml /etc/filebeat/filebeat.yml
COPY run_filebeat.sh /usr/bin/run_filebeat.sh

COPY . /opt/microservices/
RUN tar -xzf /opt/microservices/${A8_SIDECAR_RELEASE}.tar.gz -C /

EXPOSE 9080
WORKDIR /opt/microservices

ENTRYPOINT ["a8sidecar", "--config", "/opt/microservices/config.yaml"]
