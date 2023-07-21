'use strict';

const opentelemetry = require('@opentelemetry/api');
const { ParentBasedSampler,  AlwaysOnSampler, CompositePropagator, W3CTraceContextPropagator } = require( '@opentelemetry/core');
const { registerInstrumentations } = require( '@opentelemetry/instrumentation');
const { NodeTracerProvider } = require( '@opentelemetry/sdk-trace-node');
const { SimpleSpanProcessor } = require( '@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } =  require('@opentelemetry/exporter-trace-otlp-http');
const { Resource } = require( '@opentelemetry/resources');
const { SemanticResourceAttributes } = require( '@opentelemetry/semantic-conventions');
const { B3Propagator, B3InjectEncoding } = require("@opentelemetry/propagator-b3");
const { ExpressInstrumentation, ExpressLayerType } = require( '@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const axios = require("axios")


const ignoredTargets = [
  "/healthz", "/favicon.ico", "/metrics"
]

function setupTracer(serviceName){

  const provider = new NodeTracerProvider({
    resource: new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    }),
    sampler: new ParentBasedSampler({
      root: new AlwaysOnSampler()
    }),
  });

  const propagator = new CompositePropagator({
    propagators: [
      // new W3CTraceContextPropagator(), 
      new B3Propagator({injectEncoding: B3InjectEncoding.MULTI_HEADER})
    ],
  })

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
    const exporter = new OTLPTraceExporter({
      url: traceCollectorEndpoint
    });

    provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
  }

  // Initialize the OpenTelemetry APIs to use the NodeTracerProvider bindings
  provider.register({
    propagator: propagator,
  });

  return opentelemetry.trace.getTracer(serviceName);
};

module.exports = {
    setupTracer,
    startNewSpan
}


function startNewSpan(name, tracer){
  const currentSpan = opentelemetry.trace.getSpan(opentelemetry.context.active());
  const ctx = opentelemetry.trace.setSpan(
      opentelemetry.context.active(),
      currentSpan
  );
  return tracer.startSpan(name, undefined, ctx);
}
