'use strict';

const opentelemetry = require('@opentelemetry/api');
const { ParentBasedSampler,  AlwaysOnSampler, CompositePropagator, W3CTraceContextPropagator } = require( '@opentelemetry/core');
const { registerInstrumentations } = require( '@opentelemetry/instrumentation');
const { NodeTracerProvider } = require( '@opentelemetry/sdk-trace-node');
const { SimpleSpanProcessor } = require( '@opentelemetry/sdk-trace-base');
const { JaegerExporter } = require( '@opentelemetry/exporter-jaeger');
const { Resource } = require( '@opentelemetry/resources');
const { SemanticResourceAttributes } = require( '@opentelemetry/semantic-conventions');
const { SpanAttributes } = require( "@opentelemetry/api/build/src/trace/attributes");
const {B3Propagator, B3InjectEncoding} = require("@opentelemetry/propagator-b3");
const { ExpressInstrumentation, ExpressLayerType } = require( '@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const axios = require("axios")


function setupTracer(serviceName){

  const jaegerServiceEndpoint = process.env.JAEGER_SERVICE_ENDPOINT ? process.env.JAEGER_SERVICE_ENDPOINT : "http://localhost:3000"

  if(!isJeagerAvailable(jaegerServiceEndpoint)){
    console.log("Jaeger backend not installed. Skipping tracer setup...")
    return;
  }

  const exporter = new JaegerExporter({
    serviceName,
    endpoint: jaegerServiceEndpoint,
  });

  const provider = new NodeTracerProvider({
    resource: new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    }),

    sampler: new AlwaysOnSampler(),
    // sampler: b3Sampler(new ParentBasedSampler({
    //   root: new AlwaysOnSampler()
    // })),
    propagator: new CompositePropagator([new W3CTraceContextPropagator(), new B3Propagator({injectEncoding: B3InjectEncoding.MULTI_HEADER})])
  });

  registerInstrumentations({
    tracerProvider: provider,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingPaths: ignoredTargets,
        // headersToSpanAttributes: {
        //   client: {
        //     requestHeaders: ['x-b3-sampled']
        //   },
        //   server: {
        //     requestHeaders: ['x-b3-sampled']
        //   }
        // }
      }),
      new ExpressInstrumentation({
        ignoreLayersType: [ExpressLayerType.MIDDLEWARE]
      }),
    ],
  });



  provider.addSpanProcessor(new SimpleSpanProcessor(exporter));

  // Initialize the OpenTelemetry APIs to use the NodeTracerProvider bindings
  provider.register();

  return opentelemetry.trace.getTracer(serviceName);
};

// type FilterFunction = (spanName: string, spanKind: SpanKind, attributes: SpanAttributes) => boolean;

function b3Sampler(parent) {
  return {
    shouldSample(ctx, tid, spanName, spanKind, attr, links) {
      console.log("should sample? attr", attr)
      console.log("should sample? ctx", ctx)
      console.log("should sample? spanname", spanName)
      console.log("should sample? spanKind", spanKind)
      console.log("should sample? links", links)
      // if (!filterFn(spanName, spanKind, attr)) {
      //   return { decision: opentelemetry.SamplingDecision.NOT_RECORD };
      // }
      return parent.shouldSample(ctx, tid, spanName, spanKind, attr, links);
    },
    toString() {
      return `B3Sampler(${parent.toString()})`;
    }
  }
}


const ignoredTargets = [
    "/healthz", "/favicon.ico", "/metrics"
]

module.exports = {
    setupTracer,
    startNewSpan
}

async function isJeagerAvailable(endpoint){
  let jeagerAvailable = false;
  await axios(endpoint)
  .then(response => {
     console.log('resopose from jaeger ', response);  
  })
  .catch((err) => {
    // 405 is the right status code for the GET method if jaeger service exists
    // because the only allowed method is POST and usage of other methods are not allowe
    // https://github.com/jaegertracing/jaeger/blob/7872d1b07439c3f2d316065b1fd53e885b26a66f/cmd/collector/app/handler/http_handler.go#L60
    if (err.response && err.response.status === 405) {
       jeagerAvailable = true;
    }
  });

  return jeagerAvailable;
}

function startNewSpan(name, tracer){
  const currentSpan = opentelemetry.trace.getSpan(opentelemetry.context.active());
  const ctx = opentelemetry.trace.setSpan(
      opentelemetry.context.active(),
      currentSpan
  );
  return tracer.startSpan(name, undefined, ctx);
}