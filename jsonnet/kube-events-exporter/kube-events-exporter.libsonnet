local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  config+:: {
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    imageRepo: error 'must provide image repository',

    commonLabels: {
      'app.kubernetes.io/name': 'kube-events-exporter',
      'app.kubernetes.io/version': $.config.version,
      'app.kubernetes.io/component': 'events-exporter',
    },

    selectorLabels: {
      [labelName]: $.config.commonLabels[labelName]
      for labelName in std.objectFields($.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  kubeEventsExporter+:: {
    local kee = self,

    namespace:: $.config.namespace,
    version:: $.config.version,
    image:: $.config.imageRepo + ":v" + kee.version,

    commonLabels:: $.config.commonLabels,
    selectorLabels:: $.config.selectorLabels,

    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withLabels(kee.commonLabels) +
      clusterRoleBinding.mixin.metadata.withName('kube-events-exporter') +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName(kee.clusterRole.metadata.name) +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{
        kind: 'ServiceAccount',
        name: kee.serviceAccount.metadata.name,
        namespace: kee.namespace,
      }]),

    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local eventRule = policyRule.new() +
                        policyRule.withApiGroups(['']) +
                        policyRule.withResources(['events']) +
                        policyRule.withVerbs(['list', 'watch']);

      clusterRole.new() +
      clusterRole.mixin.metadata.withLabels(kee.commonLabels) +
      clusterRole.mixin.metadata.withName('kube-events-exporter') +
      clusterRole.withRules(eventRule),

    deployment:
      local deployment = k.apps.v1.deployment;
      local container = deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;

      local exporterContainer =
        container.new('kube-events-exporter', kee.image) +
        container.withPorts([
          containerPort.newNamed(8080, 'event'),
          containerPort.newNamed(8081, 'exporter'),
        ]);

      deployment.new('kube-events-exporter', 1, exporterContainer, kee.commonLabels) +
      deployment.mixin.metadata.withNamespace(kee.namespace) +
      deployment.mixin.metadata.withLabels(kee.commonLabels) +
      deployment.mixin.spec.selector.withMatchLabels(kee.selectorLabels) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.withServiceAccountName(kee.serviceAccount.metadata.name),

    podMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'PodMonitor',
      metadata: {
        labels: kee.commonLabels,
        name: 'kube-events-exporter',
        namespace: kee.namespace,
      },
      spec: {
        podMetricsEndpoints: [
          {
            port: 'event',
          },
          {
            port: 'exporter',
          },
        ],
        selector: {
          matchLabels: kee.selectorLabels,
        },
      },
    },

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('kube-events-exporter') +
      serviceAccount.mixin.metadata.withLabels(kee.commonLabels) +
      serviceAccount.mixin.metadata.withNamespace(kee.namespace),
  },
}
