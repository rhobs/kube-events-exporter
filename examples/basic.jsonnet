local lib = (import 'kube-events-exporter/kube-events-exporter.libsonnet') + {
  config+:: {
    namespace:: 'default',
    version:: std.extVar("VERSION"),
    imageRepo:: std.extVar("IMAGE_REPO"),
  },
};

{
  local kee = lib.kubeEventsExporter,

  'kube-events-exporter-cluster-role-binding': kee.clusterRoleBinding,
  'kube-events-exporter-cluster-role': kee.clusterRole,
  'kube-events-exporter-deployment': kee.deployment,
  'kube-events-exporter-pod-monitor': kee.podMonitor,
  'kube-events-exporter-service-account': kee.serviceAccount,
}
