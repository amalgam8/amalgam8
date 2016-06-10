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

FROM ubuntu:14.04

# Vulnerability Advisor stuff
RUN sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs
RUN sed -i 's/sha512/sha512 minlen=8/' /etc/pam.d/common-password

# Environment variables
ENV NGINX_PORT 6379
EXPOSE 6379

WORKDIR /opt/controller
COPY /bin/controller /opt/controller/controller
COPY /nginx/nginx.conf.tmpl /opt/controller/nginx/nginx.conf.tmpl

ENTRYPOINT ["/opt/controller/controller"]

ENV GIT_COMMIT={GIT_COMMIT} \
    IMAGE_NAME={IMAGE_NAME}
