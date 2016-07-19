#!/bin/bash
if [ "$GOPATH" == "" ]; then
    echo "Error: GOPATH needs to be set. amalgam8/sidecar needs to be installed in Gopath as well."
    exit(1)
fi

SIDECAR_PATH=${GOPATH}/src/github.com/amalgam8/sidecar

adduser --disabled-password --gecos "" nginx

apt-get -y update && apt-get -y install libreadline-dev libncurses5-dev libpcre3-dev \
    libssl-dev perl make build-essential curl wget

mkdir /var/log/nginx
mkdir /opt/a8_lualib

wget -O /tmp/ https://openresty.org/download/openresty-1.9.15.1.tar.gz

cd /tmp && \
    tar -xzvf /tmp/openresty-*.tar.gz && \
    rm -f /tmp/openresty-*.tar.gz && \
    cd /tmp/openresty-* && \
    ./configure --with-pcre-jit --with-cc-opt='-O3' --with-luajit-xcflags='-O3' --conf-path=/etc/nginx/nginx.conf --pid-path=/var/run/nginx.pid --user=nginx && \
    make && \
    make install && \
    make clean && \
    cd .. && \
    rm -rf openresty-* && \
    ln -s /usr/local/openresty/nginx/sbin/nginx /usr/local/bin/nginx && \
    ldconfig

curl -L -O https://download.elastic.co/beats/filebeat/filebeat_1.2.2_amd64.deb && \
    dpkg -i filebeat_1.2.2_amd64.deb

cp ${SIDECAR_PATH}/nginx/lua/*.lua /opt/a8_lualib/
cp ${SIDECAR_PATH}/nginx/conf/*.conf /etc/nginx/
cp ${SIDECAR_PATH}/docker/filebeat.yml /etc/filebeat/filebeat.yml

cp ${SIDECAR_PATH}/bin/a8sidecar /usr/bin/a8sidecar
