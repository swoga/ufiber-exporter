ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL org.opencontainers.image.source https://github.com/swoga/ufiber-exporter

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/ufiber-exporter /bin/ufiber-exporter
COPY example.yml /etc/ufiber-exporter/config.yml

RUN chown -R nobody:nobody /etc/ufiber-exporter

USER nobody
EXPOSE 80

ENTRYPOINT ["/bin/ufiber-exporter"]
CMD ["--config.file=/etc/ufiber-exporter/config.yml"]