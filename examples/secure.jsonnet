local lib = (import 'kube-events-exporter/kube-events-exporter.libsonnet') + {
  config+:: {
    namespace:: 'default',
    version:: std.extVar("VERSION"),
    imageRepo:: std.extVar("IMAGE_REPO"),
  },
};

local tlsCipherSuites = [
  'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305',
  'TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305',
];

{
  local kee = lib.kubeEventsExporter + 
  ((import 'kube-rbac-proxy/container.libsonnet') {
    config+:: {
      kubeRbacProxy: {
        image: 'quay.io/brancz/kube-rbac-proxy:v0.6.0',
        name: 'kube-rbac-proxy-event',
        securePortName: 'https-event',
        securePort: 8443,
        secureListenAddress: ':%d' % self.securePort,
        upstream: 'http://127.0.0.1:8080/',
        tlsCipherSuites: tlsCipherSuites
      },
    },
  }).deploymentMixin +
  ((import 'kube-rbac-proxy/container.libsonnet') {
    config+:: {
      kubeRbacProxy: {
        image: 'quay.io/brancz/kube-rbac-proxy:v0.6.0',
        name: 'kube-rbac-proxy-exporter',
        securePortName: 'https-exporter',
        securePort: 9443,
        secureListenAddress: ':%d' % self.securePort,
        upstream: 'http://127.0.0.1:8081/',
        tlsCipherSuites: tlsCipherSuites
      },
    },
  }).deploymentMixin +
  {
    service+: {
      spec+: {
        ports: [
          {
            name: port.name,
            port: port.port,
            targetPort: port.name,
          }
          for port in [
            { name: 'https-event', port: 8443 },
            { name: 'https-exporter', port: 9443 },
          ]
        ],
      },
    },
    serviceMonitor+: {
      spec+: {
        endpoints: [
          {
            port: port,
            scheme: 'https',
            bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            tlsConfig: {
              insecureSkipVerify: true,
            },
          } 
          for port in ['https-event', 'https-exporter']
        ],
      },
    },
    clusterRole+: {
      rules+: [
        {
          apiGroups: ['authentication.k8s.io'],
          resources: ['tokenreviews'],
          verbs: ['create'],
        },
        {
          apiGroups: ['authorization.k8s.io'],
          resources: ['subjectaccessreviews'],
          verbs: ['create'],
        },
      ],
    },
  },

  'kube-events-exporter-cluster-role-binding': kee.clusterRoleBinding,
  'kube-events-exporter-cluster-role': kee.clusterRole,
  'kube-events-exporter-deployment': kee.deployment,
  'kube-events-exporter-service-account': kee.serviceAccount,
  'kube-events-exporter-service': kee.service,
  'kube-events-exporter-service-monitor': kee.serviceMonitor,
}
