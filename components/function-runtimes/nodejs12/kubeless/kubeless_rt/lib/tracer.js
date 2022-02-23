'use strict';

const axios = require('axios');
const opentelemetry = require('@opentelemetry/api');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { NodeTracerProvider } = require('@opentelemetry/node');
const { SimpleSpanProcessor } = require('@opentelemetry/tracing');
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { B3Propagator,B3InjectEncoding } = require("@opentelemetry/propagator-b3");

function is_jaeger_available(endpoint) {
  let res = await axios(endpoint);

  // 405 is the right status code for the GET method if jaeger service exists
  if (res.status == 405) {
    return true;
  }
  return false;
}

module.exports = (serviceName, endpoint) => {
  if (is_jaeger_available(endpoint) == false) {
    return null;
  }


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
