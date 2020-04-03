FROM busybox

ADD kube-events-exporter /

ENTRYPOINT ["/kube-events-exporter"]

EXPOSE 8080
