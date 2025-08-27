export default [
  { text: 'Istio Module', link: './istio/user/README', collapsed: true },
  { text: 'Istio Sidecar Proxies', link: './istio/user/00-00-istio-sidecar-proxies' },
  { text: 'Istio Version', link: './istio/user/00-10-istio-version' },
  { text: 'Default Istio Configuration', link: './istio/user/00-15-overview-istio-setup' },
  { text: 'Istio Custom Resource', link: './istio/user/04-00-istio-custom-resource' },
  { text: 'Tutorials', link: './istio/user/tutorials/README', collapsed: true, items: [
    { text: 'Enable Istio Sidecar Injection', link: './istio/user/tutorials/01-40-enable-sidecar-injection' },
    { text: 'Expose Workloads Using oauth2-proxy', link: './istio/user/tutorials/01-10-external-authorization-provider' },
    { text: 'Expose Workloads Using Gateway API', link: './istio/user/tutorials/01-20-expose-httbin-gateway-api' },
    { text: 'Expose a TCP Service Using Gateway API Aplha Support', link: './istio/user/tutorials/01-30-expose-tcp-gateway-api' },
    { text: 'Exposing Workloads Using Istio VirtualService', link: './istio/user/tutorials/01-65-expose-workload-vs' },
    { text: 'Configure Istio Access Logs', link: './istio/user/tutorials/01-45-enable-istio-access-logs' },
    { text: 'Send Requests Using Istio Egress Gateway', link: './istio/user/tutorials/01-50-send-requests-using-egress' },
    { text: 'Send mTLS Requests Using Istio Egress Gateway', link: './istio/user/tutorials/01-55-send-requests-using-egress-and-mtls' },
    { text: 'Migrate From ELB to NLB', link: './istio/user/tutorials/01-60-enable-nlb-load-balancer' }
    ] },
  { text: 'Technical Reference', link: './istio/user/technical-reference/README', collapsed: true, items: [
    { text: 'Istio Controller Parameters', link: './istio/user/technical-reference/05-00-istio-controller-parameters' },
    { text: 'Istio Controller RBAC Configuration', link: './istio/user/technical-reference/05-10-istio-controller-rbac' }
    ] },
  { text: 'Troubleshooting', link: './istio/user/troubleshooting/README', collapsed: true, items: [
    { text: 'Network Connectivity - Basic Diagnostics', link: './istio/user/troubleshooting/03-00-network-connectivity' },
    { text: 'Connection Refused Errors', link: './istio/user/troubleshooting/03-20-connection-refused' },
    { text: 'No Access Error', link: './istio/user/troubleshooting/03-10-503-no-access' },
    { text: 'Istio Sidecar Injection Issues', link: './istio/user/troubleshooting/03-30-istio-no-sidecar' },
    { text: 'Reverting the Istio module\'s deletion', link: './istio/user/troubleshooting/03-50-recovering-from-unintentional-istio-removal' },
    { text: 'SAP HANA Database Connection Issues', link: './istio/user/troubleshooting/03-80-cannot-connect-to-hana-db' },
    { text: 'Not Found Error', link: './istio/user/troubleshooting/03-60-404-on-istio-gateway' },
    { text: 'Incompatible Istio Sidecar', link: './istio/user/troubleshooting/03-40-incompatible-istio-sidecar-version' },
    { text: 'Modified Istio Resources', link: './istio/user/troubleshooting/03-70-reconciliation-fails-on-istio-install' },
    { text: 'Istio Cannot Verify an HTTPS Certificate', link: './istio/user/troubleshooting/03-90-istio-cert-unknown' }
    ] }
];
