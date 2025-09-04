export default [
  { text: 'Telemetry Manager', link: './01-manager' },
  { text: 'Gateways', link: './gateways' },
  { text: 'Application Logs (Fluent Bit)', link: './02-logs' },
  { text: 'Application Logs (OTLP)', link: './logs' },
  { text: 'Traces', link: './03-traces' },
  { text: 'Metrics', link: './04-metrics' },
  { text: 'Integration Guides', link: './integration/README', collapsed: true, items: [
    { text: 'SAP Cloud Logging', link: './integration/sap-cloud-logging/README' },
    { text: 'Dynatrace', link: './integration/dynatrace/README' },
    { text: 'Prometheus', link: './integration/prometheus/README' },
    { text: 'Loki', link: './integration/loki/README' },
    { text: 'Jaeger', link: './integration/jaeger/README' },
    { text: 'Amazon CloudWatch', link: './integration/aws-cloudwatch/README' },
    { text: 'OpenTelemetry Demo App', link: './integration/opentelemetry-demo/README' },
    { text: 'Sample App', link: './integration/sample-app/README' }
  ]},
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'Telemetry', link: './resources/01-telemetry' },
    { text: 'LogPipeline', link: './resources/02-logpipeline' },
    { text: 'TracePipeline', link: './resources/04-tracepipeline' },
    { text: 'MetricPipeline', link: './resources/05-metricpipeline' }
  ]}
];
