'use strict';

const { SpanStatusCode } = require("@opentelemetry/api/build/src/trace/status");

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

function handleTimeOut(req, res, next){
  const timeout = Number(process.env.FUNC_TIMEOUT || '180');
  res.setTimeout(timeout*1000, function(){
          res.sendStatus(408);
      });
  next();
}

const isFunction = (func) => {
  return func && func.constructor && func.call && func.apply;
};

const isPromise = (promise) => {
  return typeof promise.then == "function"
}


function handleError(err, span, sendResponse) {
    console.error(err);
    const errTxt = resolveErrorMsg(err);
    span.setStatus({ code: SpanStatusCode.ERROR, message: errTxt });
    span.setAttribute("error", errTxt);
    sendResponse(errTxt, 500);
}

function resolveErrorMsg(err) {
    let errText
    if (typeof err == "string") {
        errText = err
    } else {
        errText = "Internal server error"
    }
    return errText
}

module.exports = {
  configureGracefulShutdown,
  handleTimeOut,
  isFunction,
  isPromise, 
  handleError
};