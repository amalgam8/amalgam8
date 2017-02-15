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

FROM amalgam8/envoy:latest

ARG FILEBEAT_VERSION="5.1.1"
RUN set -ex \
    && apt-get -y install curl \
    && curl -fsSL https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-${FILEBEAT_VERSION}-linux-x86_64.tar.gz -o /tmp/filebeat.tar.gz \
    && cd /tmp \
    && tar -xzf filebeat.tar.gz \
    \
    && cd filebeat-* \
    && cp filebeat /bin \
    \
    && rm -rf /tmp/filebeat*

ADD bin/a8sidecar /usr/bin/a8sidecar
ADD docker/filebeat.yml /etc/filebeat/filebeat.yml
ADD docker/run_filebeat.sh /usr/bin/run_filebeat.sh

ENTRYPOINT ["/usr/bin/a8sidecar"]

EXPOSE 6379
