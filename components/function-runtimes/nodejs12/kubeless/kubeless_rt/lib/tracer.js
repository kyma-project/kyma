'use strict';

const opentelemetry = require('@opentelemetry/api');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { NodeTracerProvider } = require('@opentelemetry/node');
const { SimpleSpanProcessor } = require('@opentelemetry/tracing');
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { B3Propagator,B3InjectEncoding } = require("@opentelemetry/propagator-b3");

module.exports = (serviceName, endpoint) => {
  const provider = new NodeTracerProvider();

  registerInstrumentations({
    tracerProvider: provider,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingPaths: [
          "/healthz",
          "/metrics"
        ],
        requestHook: (span, request) => {
          span.updateName(serviceName.concat('-service'))
        }
      }),
    ],
  });

  const exporter = new JaegerExporter({
    endpoint,
    serviceName,
  });

  provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

  provider.register({
    propagator: new B3Propagator({ injectEncoding: B3InjectEncoding.MULTI_HEADER }),
  });

  return opentelemetry.trace.getTracer(serviceName.concat('-tracer'),'0.0.1');
};
