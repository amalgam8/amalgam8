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

FROM websphere-liberty:latest

ARG A8_SIDECAR_RELEASE

# Install Filebeat
ARG FILEBEAT_VERSION="5.1.1"
RUN set -ex \
    \
    && apt-get update -y \
    && apt-get install curl -y \
    \
    && curl -fsSL https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-${FILEBEAT_VERSION}-linux-x86_64.tar.gz -o /tmp/filebeat.tar.gz \
    && cd /tmp \
    && tar -xzf filebeat.tar.gz \
    \
    && cd filebeat-* \
    && cp filebeat /bin \
    \
    && rm -rf /tmp/filebeat*

COPY filebeat.yml /etc/filebeat/filebeat.yml
COPY config.yaml /opt/microservices/config.yaml
COPY run_filebeat.sh /usr/bin/run_filebeat.sh
COPY ${A8_SIDECAR_RELEASE}.tar.gz /opt/microservices/

RUN tar -xzf /opt/microservices/${A8_SIDECAR_RELEASE}.tar.gz -C /

ENV SERVERDIRNAME reviews

ADD ./servers/LibertyProjectServer /opt/ibm/wlp/usr/servers/defaultServer/

RUN /opt/ibm/wlp/bin/installUtility install  --acceptLicense /opt/ibm/wlp/usr/servers/defaultServer/server.xml

ARG service_version
ARG enable_ratings
ARG star_color
ENV SERVICE_VERSION ${service_version:-v1}
ENV ENABLE_RATINGS ${enable_ratings:-false}
ENV STAR_COLOR ${star_color:-black}
ENV PROXY_SERVICE http://127.0.0.1:6379

CMD [ "a8sidecar", "--config", "/opt/microservices/config.yaml" ]
