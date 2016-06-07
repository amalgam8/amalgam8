FROM scratch

WORKDIR /opt/registry
COPY bin/registry       /opt/registry/
COPY messages        /opt/registry/locales

EXPOSE 8080

ENTRYPOINT [ "/opt/registry/registry" ]
