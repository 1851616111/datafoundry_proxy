FROM alpine

COPY oauth-proxy proxyinit.sh /usr/bin

RUN chmod +x /usr/bin/oauth-proxy /usr/bin/proxyinit.sh && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE  9090

ENV DATAFOUNDRY_APISERVER_ADDR lab.dataos.io:8443

ENTRYPOINT ["proxyinit.sh"]

