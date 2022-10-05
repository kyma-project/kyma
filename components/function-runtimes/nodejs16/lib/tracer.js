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

  const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT;

  if(!isTraceCollectorAvailable(traceCollectorEndpoint)){
    console.log("Trace collector not installed. Skipping tracer setup...")
    return;
  }

  const exporter = new OTLPTraceExporter({
    url: traceCollectorEndpoint
  });

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
      new W3CTraceContextPropagator(), 
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

  provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

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

async function isTraceCollectorAvailable(endpoint){
  let traceCollectorAvailable = false;
  console.log('checking availibility of trace collector', endpoint);  
  await axios(endpoint)
  .then(response => {
     console.log('response from trace collector', response);  
  })
  .catch((err) => {
    if (err.response && err.response.status === 405) {
      // TODO: resolve dependencies via serverless operator
      // 405 is the right status code for the GET method if jaeger service exists
      // because the only allowed method is POST and usage of other methods are not allowe
      // https://github.com/jaegertracing/jaeger/blob/7872d1b07439c3f2d316065b1fd53e885b26a66f/cmd/collector/app/handler/http_handler.go#L60
      traceCollectorAvailable = true;
    }
  });
  return traceCollectorAvailable;
}

function startNewSpan(name, tracer){
  const currentSpan = opentelemetry.trace.getSpan(opentelemetry.context.active());
  const ctx = opentelemetry.trace.setSpan(
      opentelemetry.context.active(),
      currentSpan
  );
  return tracer.startSpan(name, undefined, ctx);
}