'use strict';

const axios = require('axios');

const opentelemetry = require('@opentelemetry/api');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { NodeTracerProvider } = require('@opentelemetry/node');
const { SimpleSpanProcessor } = require('@opentelemetry/tracing');
const { OTLPTraceExporter } =  require('@opentelemetry/exporter-trace-otlp-grpc');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { B3Propagator,B3InjectEncoding } = require("@opentelemetry/propagator-b3");
const  {NoopTracerProvider} = require('@opentelemetry/api')
const {ExpressInstrumentation} = require("@opentelemetry/instrumentation-express");
const { Resource } = require( '@opentelemetry/resources');

const TRACER_SAMPLE_HEADER= "x-b3-sampled"

class ServerlessTracerProvider {
  constructor(serviceName, endpoint) {
    this.noopTracerProvider = new NoopTracerProvider().getTracer(serviceName.concat('-tracer'),'0.0.1');
    this.tracerProvider = new NoopTracerProvider().getTracer(serviceName.concat('-tracer'),'0.0.1')
    axios(endpoint)
        .catch((err) => {
          // 405 is the right status code for the GET method if jaeger service exists
          // because the only allowed method is POST and usage of other methods are not allowe
          // https://github.com/jaegertracing/jaeger/blob/7872d1b07439c3f2d316065b1fd53e885b26a66f/cmd/collector/app/handler/http_handler.go#L60
          if (err.response && err.response.status === 405) {
            this.tracerProvider = getTracerProvider(serviceName, endpoint)
          }
        });
  }

  getTracer(req) {
      if (req[TRACER_SAMPLE_HEADER] === "1") {
          return this.tracerProvider
      }
      return this.noopTracerProvider
  }
}

function getTracerProvider (serviceName, endpoint) {
  const provider = new NodeTracerProvider({
    resource: new Resource({
      ["service.name"]: serviceName,
    })
  });

  registerInstrumentations({
    tracerProvider: provider,
    instrumentations: [
      new ExpressInstrumentation(),
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

  const collectorOptions = {
    url: endpoint
  }

  const exporter = new OTLPTraceExporter(collectorOptions);

  provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

  provider.register({
    propagator: new B3Propagator({ injectEncoding: B3InjectEncoding.MULTI_HEADER }),
  });

  return opentelemetry.trace.getTracer(serviceName.concat('-tracer'),'0.0.1');
}

module.exports = {
  ServerlessTracerProvider
}
