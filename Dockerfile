FROM alpine

COPY oauth-proxy /usr/bin

RUN chmod +x /usr/bin/oauth-proxy && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && \
    echo search home > /etc/resolv.conf && \
    echo nameserver 223.5.5.5 >> /etc/resolv.conf && \
    echo nameserver 223.6.6.6 >> /etc/resolv.conf 

EXPOSE  9090

ENV DATAFOUNDRY_APISERVER_ADDR lab.asiainfodata.com:8443

ENTRYPOINT ["oauth-proxy"]

