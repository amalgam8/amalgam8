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
