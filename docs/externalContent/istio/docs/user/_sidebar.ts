export default [
  { text: 'Istio Sidecar Proxies', link: './00-00-istio-sidecar-proxies' },
  { text: 'Istio Version', link: './00-10-istio-version' },
  { text: 'Default Istio Configuration', link: './00-15-overview-istio-setup' },
  { text: 'Istio Custom Resource', link: './04-00-istio-custom-resource' },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Enable Istio Sidecar Injection', link: './tutorials/01-40-enable-sidecar-injection' },
    { text: 'Expose Workloads Using oauth2-proxy', link: './tutorials/01-10-external-authorization-provider' },
    { text: 'Expose Workloads Using Gateway API', link: './tutorials/01-20-expose-httbin-gateway-api' },
    { text: 'Expose a TCP Service Using Gateway API Aplha Support', link: './tutorials/01-30-expose-tcp-gateway-api' },
    { text: 'Exposing Workloads Using Istio VirtualService', link: './tutorials/01-65-expose-workload-vs' },
    { text: 'Configure Istio Access Logs', link: './tutorials/01-45-enable-istio-access-logs' },
    { text: 'Send Requests Using Istio Egress Gateway', link: './tutorials/01-50-send-requests-using-egress' },
    { text: 'Send mTLS Requests Using Istio Egress Gateway', link: './tutorials/01-55-send-requests-using-egress-and-mtls' },
    { text: 'Migrate From ELB to NLB', link: './tutorials/01-60-enable-nlb-load-balancer' }
    ] },
  { text: 'Technical Reference', link: './technical-reference/README', collapsed: true, items: [
    { text: 'Istio Controller Parameters', link: './technical-reference/05-00-istio-controller-parameters' },
    { text: 'Istio Controller RBAC Configuration', link: './technical-reference/05-10-istio-controller-rbac' }
    ] },
  { text: 'Troubleshooting', link: './troubleshooting/README', collapsed: true, items: [
    { text: 'Network Connectivity - Basic Diagnostics', link: './troubleshooting/03-00-network-connectivity' },
    { text: 'Connection Refused Errors', link: './troubleshooting/03-20-connection-refused' },
    { text: 'No Access Error', link: './troubleshooting/03-10-503-no-access' },
    { text: 'Istio Sidecar Injection Issues', link: './troubleshooting/03-30-istio-no-sidecar' },
    { text: 'Reverting the Istio module\'s deletion', link: './troubleshooting/03-50-recovering-from-unintentional-istio-removal' },
    { text: 'SAP HANA Database Connection Issues', link: './troubleshooting/03-80-cannot-connect-to-hana-db' },
    { text: 'Not Found Error', link: './troubleshooting/03-60-404-on-istio-gateway' },
    { text: 'Incompatible Istio Sidecar', link: './troubleshooting/03-40-incompatible-istio-sidecar-version' },
    { text: 'Modified Istio Resources', link: './troubleshooting/03-70-reconciliation-fails-on-istio-install' },
    { text: 'Istio Cannot Verify an HTTPS Certificate', link: './troubleshooting/03-90-istio-cert-unknown' }
    ] }
];
