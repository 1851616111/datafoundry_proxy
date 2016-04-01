FROM alpine

COPY oauthServer /usr/bin

RUN chmod +x /usr/bin/oauthServer && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE  9090

ENV DATAFOUNDRY_APISERVER_ADDR lab.asiainfodata.com:8443

ENTRYPOINT ["oauthServer"]

