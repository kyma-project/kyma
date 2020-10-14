'use strict';

const fs = require('fs');

function readDependencies(pkgFile) {
  try {
    const data = JSON.parse(fs.readFileSync(pkgFile));
    const deps = data.dependencies;
    return (deps && typeof deps === 'object') ? Object.getOwnPropertyNames(deps) : [];
  } catch(e) {
    return [];
  }
}

function prepareStatistics(label, promClient) {
  const timeHistogram = new promClient.Histogram({
    name: 'function_duration_seconds',
    help: 'Duration of user function in seconds',
    labelNames: [label],
  });
  const callsCounter = new promClient.Counter({
    name: 'function_calls_total',
    help: 'Number of calls to user function',
    labelNames: [label],
  });
  const errorsCounter = new promClient.Counter({
    name: 'function_failures_total',
    help: 'Number of exceptions in user function',
    labelNames: [label],
  });
  return {
    timeHistogram,
    callsCounter,
    errorsCounter,
  };
}

function routeLivenessProbe(expressApp) {
  expressApp.get('/healthz', (req, res) => {
    res.status(200).send('OK');
  });
}

function routeMetrics(expressApp, promClient) {
  expressApp.get('/metrics', (req, res) => {
    res.status(200);
    res.type(promClient.register.contentType);
    res.send(promClient.register.metrics());
  });
}

function configureGracefulShutdown(server) {
  let nextConnectionId = 0;
  const connections = [];
  let terminating = false;

  server.on('connection', connection => {
    const connectionId = nextConnectionId++;
    connection.$$isIdle = true;
    connections[connectionId] = connection;
    connection.on('close', () => delete connections[connectionId]);
  });

  server.on('request', (request, response) => {
    const connection = request.connection;
    connection.$$isIdle = false;

    response.on('finish', () => {
      connection.$$isIdle = true;
      if (terminating) {
        connection.destroy();
      }
    });
  });

  const handleShutdown = () => {
    console.log("Shutting down..");

    terminating = true;
    server.close(() => console.log("Server stopped"));

    for (const connectionId in connections) {
      if (connections.hasOwnProperty(connectionId)) {
        const connection = connections[connectionId];
        if (connection.$$isIdle) {
          connection.destroy();
        }
      }
    }
  };

  process.on('SIGINT', handleShutdown);
  process.on('SIGTERM', handleShutdown);
}

module.exports = {
  readDependencies,
  prepareStatistics,
  routeLivenessProbe,
  routeMetrics,
  configureGracefulShutdown
};
