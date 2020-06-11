ARG ARCH="amd64"
ARG OS="linux"

FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

ADD kube-events-exporter /

ENTRYPOINT ["/kube-events-exporter"]

EXPOSE 8080 8081
