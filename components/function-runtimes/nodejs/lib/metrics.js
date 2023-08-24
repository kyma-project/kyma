const opentelemetry = require('@opentelemetry/api');
const { MeterProvider } = require('@opentelemetry/sdk-metrics');
const { PrometheusExporter } = require('@opentelemetry/exporter-prometheus');
const { Resource } = require( '@opentelemetry/resources');
const { SemanticResourceAttributes } = require( '@opentelemetry/semantic-conventions');

let exporter;

function setupMetrics(serviceName){

    const resource = new Resource({
        [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    });


    const myServiceMeterProvider = new MeterProvider({
    resource,
    });

    exporter = new PrometheusExporter({ preventServerStart: true})

    myServiceMeterProvider.addMetricReader(exporter);

    opentelemetry.metrics.setGlobalMeterProvider(myServiceMeterProvider);

}

function createFunctionCallsTotalCounter(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_calls_total',{
    description: 'Number of calls to user function',
  }); 
}
  
  
function createFunctionFailuresTotalCounter(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_failures_total',{
    description: 'Number of exceptions in user function',
  });  
}

function createFunctionDurationHistogram(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createHistogram("function_duration_miliseconds",{
    description: 'Duration of user function in miliseconds',
  });  
}

const getMetrics = (req, res) => {
  exporter.getMetricsRequestHandler(req, res);
};

module.exports = {
    setupMetrics,
    createFunctionCallsTotalCounter,
    createFunctionFailuresTotalCounter,
    createFunctionDurationHistogram,
    getMetrics,
}