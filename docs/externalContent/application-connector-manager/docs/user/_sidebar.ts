export default [
  { text: 'Architecture', link: './00-10-architecture' },
  { text: 'Technical Reference', link: './technical-reference/README', collapsed: true, items: [
    { text: 'Istio Gateway', link: './technical-reference/02-10-istio-gateway' },
    { text: 'Application Gateway', link: './technical-reference/02-20-application-gateway' },
    { text: 'Application Connectivity Validator', link: './technical-reference/02-30-application-connectivity-validator' },
    { text: 'Runtime Agent', link: './technical-reference/runtime-agent/README', collapsed: true, items: [
      { text: 'UCL Connection', link: './technical-reference/runtime-agent/03-10-ucl-connection' },
      { text: 'Configuring Runtime', link: './technical-reference/runtime-agent/03-20-configuring-runtime' },
      { text: 'Tutorials', link: './technical-reference/runtime-agent/tutorials/README', collapsed: true, items: [
        { text: 'Establish a Secure Connection with UCL', link: './technical-reference/runtime-agent/tutorials/01-60-establish-secure-connection-with-compass' },
        { text: 'Maintain a Secure Connection with UCL', link: './technical-reference/runtime-agent/tutorials/01-70-maintain-secure-connection-with-compass' },
        { text: 'Revoke a Client Certificate (RA)', link: './technical-reference/runtime-agent/tutorials/01-80-revoke-client-certificate' },
        { text: 'Configure Runtime Agent with UCL', link: './technical-reference/runtime-agent/tutorials/01-90-configure-runtime-agent-with-compass' },
        { text: 'Reconnect Runtime Agent with UCL', link: './technical-reference/runtime-agent/tutorials/01-100-reconnect-runtime-agent-with-compass' }
        ] }
      ] }
    ] },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'Application CR', link: './resources/04-10-application' },
    { text: 'Application Connector CR', link: './resources/04-30-application-connector' },
    { text: 'CompassConnection CR', link: './resources/04-20-compassconnection' }
    ] },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Integrate an External System with Kyma', link: './tutorials/01-00-integrate-external-system' },
    { text: 'Create an Application CR', link: './tutorials/01-10-create-application' },
    { text: 'Register a Service', link: './tutorials/01-20-register-manage-services' },
    { text: 'Register a Secured API', link: './tutorials/01-30-register-secured-api' },
    { text: 'Disable TLS Validation', link: './tutorials/01-50-disable-tls-certificate-verification' },
    { text: 'Call a Registered External Service from Kyma', link: './tutorials/01-40-call-registered-service-from-kyma' }
    ] }
];
