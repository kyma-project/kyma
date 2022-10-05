'use strict';

const axios = require('axios');

const opentelemetry = require('@opentelemetry/api');
const { registerInstrumentations } = require( '@opentelemetry/instrumentation');
const { NodeTracerProvider } = require( '@opentelemetry/sdk-trace-node');
const { SimpleSpanProcessor, SamplingDecision } = require( '@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } =  require('@opentelemetry/exporter-trace-otlp-http');
const { Resource } = require( '@opentelemetry/resources');
const { B3Propagator, B3InjectEncoding } = require("@opentelemetry/propagator-b3");
const { ExpressInstrumentation, ExpressLayerType } = require( '@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');

const ignoredTargets = [
  "/healthz", "/favicon.ico", "/metrics"
]

function getTracer(serviceName) {
 
  const provider = new NodeTracerProvider({
    resource: new Resource({
      ["service.name"]: serviceName,
    }),
    sampler: {
      shouldSample(context) {
        const parentSpanContext = opentelemetry.trace.getSpanContext(context)
        if (parentSpanContext && (parentSpanContext.traceFlags & opentelemetry.TraceFlags.SAMPLED)) {
          return {
            decision: SamplingDecision.RECORD_AND_SAMPLED
          };  
        } else {
          return {
            decision: SamplingDecision.NOT_RECORD
          };
        }
      },
      toString() {
          return 'KymaFunctionSampler';
      }
    }
  });

  registerInstrumentations({
    tracerProvider: provider,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingPaths: ignoredTargets,
      }),
      new ExpressInstrumentation({
        ignoreLayersType: [ExpressLayerType.MIDDLEWARE]
      }),
    ],
  });

  const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT;

  if(traceCollectorEndpoint){
    axios(traceCollectorEndpoint).catch((err) => {
      if (err.response && err.response.status === 405) {
        // TODO: resolve dependencies via serverless operator
        // 405 is the right status code for the GET method if jaeger service exists
        // because the only allowed method is POST and usage of other methods are not allowe
        // https://github.com/jaegertracing/jaeger/blob/7872d1b07439c3f2d316065b1fd53e885b26a66f/cmd/collector/app/handler/http_handler.go#L60
        const exporter = new OTLPTraceExporter({
          url: traceCollectorEndpoint
        });
    
        provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
      }
    });
  }

  // Initialize the OpenTelemetry APIs to use the NodeTracerProvider bindings
  provider.register({
    propagator: new B3Propagator({injectEncoding: B3InjectEncoding.MULTI_HEADER}),
  });

  return opentelemetry.trace.getTracer(serviceName);
}

module.exports = {
  getTracer
}