#!/bin/bash
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
BUILD_BASE=${SCRIPTDIR}/openresty
DEST=${SCRIPTDIR}/dist
OS="linux"
ARCH="amd64"
RESTY_VERSION="1.11.2.1"
RESTY_LUAROCKS_VERSION="2.3.0"
RESTY_OPENSSL_VERSION="1.0.2h"
RESTY_PCRE_VERSION="8.39"
RESTY_J="2"
RESTY_CONFIG_OPTIONS="\
    --with-file-aio \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_dav_module \
    --with-http_flv_module \
    --with-http_geoip_module=dynamic \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_image_filter_module=dynamic \
    --with-http_mp4_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-http_xslt_module=dynamic \
    --with-ipv6 \
    --with-mail \
    --with-mail_ssl_module \
    --with-md5-asm \
    --with-pcre-jit \
    --with-sha1-asm \
    --with-stream \
    --with-stream_ssl_module \
    --with-threads \
    --with-luajit-xcflags='-O3' \
    --with-cc-opt='-O3' \
    --sbin-path=/usr/sbin/nginx \
    --error-log-path=/var/log/nginx/error.log \
    --pid-path=/var/run/nginx.pid \
    --http-log-path=/var/log/nginx/access.log \
    --conf-path=/etc/nginx/nginx.conf
    "
_RESTY_CONFIG_DEPS="--with-openssl=${BUILD_BASE}/openssl-${RESTY_OPENSSL_VERSION} --with-pcre=${BUILD_BASE}/pcre-${RESTY_PCRE_VERSION}"
DEBIAN_FRONTEND=noninteractive apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y rubygems-integration ruby-dev gcc make && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    libgd-dev \
    libgeoip-dev \
    libncurses5-dev \
    libperl-dev \
    libreadline-dev \
    libxslt1-dev \
    make \
    perl \
    unzip \
    zlib1g-dev \
    && mkdir -p ${BUILD_BASE} \
    && cd ${BUILD_BASE} \
    && curl -fSL https://www.openssl.org/source/openssl-${RESTY_OPENSSL_VERSION}.tar.gz -o openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && tar xzf openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && curl -fSL https://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-${RESTY_PCRE_VERSION}.tar.gz -o pcre-${RESTY_PCRE_VERSION}.tar.gz \
    && tar xzf pcre-${RESTY_PCRE_VERSION}.tar.gz \
    && curl -fSL https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz -o openresty-${RESTY_VERSION}.tar.gz \
    && tar xzf openresty-${RESTY_VERSION}.tar.gz \
    && cd ${BUILD_BASE}/openresty-${RESTY_VERSION} \
    && ./configure -j${RESTY_J} ${_RESTY_CONFIG_DEPS} ${RESTY_CONFIG_OPTIONS} \
    && make -j${RESTY_J} \
    && mkdir ${DEST} \
    && DESTDIR=${DEST} make install \
    && cd ${SCRIPTDIR} \
	&& tar -C ${DEST} -czf ${SCRIPTDIR}/openresty-${RESTY_VERSION}-bin-${OS}-${ARCH}.tar.gz --transform 's:^./::' .


    # && make -j${RESTY_J} install \
    # && cd /tmp \
    # && rm -rf \
    #     openssl-${RESTY_OPENSSL_VERSION} \
    #     openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    #     openresty-${RESTY_VERSION}.tar.gz openresty-${RESTY_VERSION} \
    #     pcre-${RESTY_PCRE_VERSION}.tar.gz pcre-${RESTY_PCRE_VERSION} \
    # && curl -fSL http://luarocks.org/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz -o luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    # && tar xzf luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    # && cd luarocks-${RESTY_LUAROCKS_VERSION} \
    # && ./configure \
    #     --prefix=/usr/local/openresty/luajit \
    #     --with-lua=/usr/local/openresty/luajit \
    #     --lua-suffix=jit-2.1.0-beta2 \
    #     --with-lua-include=/usr/local/openresty/luajit/include/luajit-2.1 \
    # && make build \
    # && make install \
    # && cd /tmp \
    # && rm -rf luarocks-${RESTY_LUAROCKS_VERSION} luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz \
    # && DEBIAN_FRONTEND=noninteractive apt-get autoremove -y \

    # && fpm -s dir -t deb -n openresty -v ${RESTY_VERSION} -C ${BUILD_BASE}/build/root -p openresty_VERSION_ARCH.deb \
    #     --description "A high performance web server and a reverse proxy server with Openresty extensions"  --url 'http://openresty.org/' \
    #     --category httpd -d libncurses5 -d libreadline6 -d ca-certificates \
    #     -m 'Amalgam8 Team' --vendor "Amalgam8 Team" \
    #     --replaces 'nginx-full' --provides 'nginx-full' \
    #     --conflicts 'nginx-full' --replaces 'nginx-common' --provides 'nginx-common' --conflicts 'nginx-common' \
    #     --provides 'nginx-core' --conflicts 'nginx-core' \
    #     --deb-build-depends build-essential
