##
# BEGIN Proxy Base Image
##
FROM ubuntu:14.04

# Vulnerability Advisor stuff
RUN sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs
RUN sed -i 's/sha512/sha512 minlen=8/' /etc/pam.d/common-password

# Install necessary packages
RUN sudo -s && \
    apt-get update -y && \
    apt-get install -y wget apt-transport-https && \
    wget -O - https://downloads.opvis.bluemix.net:5443/client/IBM_Logmet_repo_install.sh | bash && \
    apt-get install -y --only-upgrade libpng12-0 libexpat1 libpcre3 libsqlite3-0 dpkg libgnutls26 libssl1.0.0 && \
    apt-get update -y && \
    apt-get install -y --force-yes collectd-write-mtlumberjack ca-certificates && \
    apt-get upgrade logrotate -y

ENV DEBIAN_FRONTEND=noninteractive

##
# END Proxy Base Image
##

# Environment variables
ENV NGINX_PORT 6379
EXPOSE 6379

ENTRYPOINT ["/controller"]

COPY /nginx/nginx.conf.tmpl /nginx/nginx.conf.tmpl
COPY /controller /

ENV GIT_COMMIT={GIT_COMMIT} \
    IMAGE_NAME={IMAGE_NAME}
