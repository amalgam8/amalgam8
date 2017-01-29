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

FROM ruby:2.3

ARG A8_SIDECAR_RELEASE

RUN mkdir -p /opt/microservices
COPY details.rb /opt/microservices/

COPY ${A8_SIDECAR_RELEASE}.tar.gz /opt/microservices/
RUN tar -xzf /opt/microservices/${A8_SIDECAR_RELEASE}.tar.gz -C /

EXPOSE 9080
WORKDIR /opt/microservices
ENTRYPOINT ["a8sidecar", "ruby", "details.rb", "9080"]
