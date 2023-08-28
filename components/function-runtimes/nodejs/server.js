"use strict";
const ce = require('./lib/ce');
const helper = require('./lib/helper');
const bodyParser = require('body-parser');
const process = require("process");
const morgan = require("morgan");

const { setupTracer, startNewSpan } = require('./lib/tracer')
const { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter  } = require('./lib/metrics')


// To catch unhandled exceptions thrown by user code async callbacks,
// these exceptions cannot be catched by try-catch in user function invocation code below
process.on("uncaughtException", (err) => {
    console.error(`Caught exception: ${err}`);
});

const podName = process.env.HOSTNAME || "";
const serviceNamespace = process.env.SERVICE_NAMESPACE || "";
let serviceName = podName.substring(0, podName.lastIndexOf("-"));
serviceName = serviceName.substring(0, serviceName.lastIndexOf("-"));
const defaultFunctioneName = serviceName.substring(0, serviceName.lastIndexOf("-"));
const functionName = process.env.FUNC_NAME || defaultFunctioneName;
const bodySizeLimit = Number(process.env.REQ_MB_LIMIT || '1');
const funcPort = Number(process.env.FUNC_PORT || '8080');

const otelServiceName = [serviceName, serviceNamespace].join('.')
const tracer = setupTracer(otelServiceName);
setupMetrics(otelServiceName);

const callsTotalCounter = createFunctionCallsTotalCounter(otelServiceName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(otelServiceName);
const durationHistogram = createFunctionDurationHistogram(otelServiceName);

//require express must be called AFTER tracer was setup!!!!!!
const express = require("express");
const app = express();


// User function.  Starts out undefined.
let userFunction;

const loadFunction = (modulepath, funcname) => {
    // Read and load the code. It's placed there securely by the fission runtime.
    try {
        let startTime = process.hrtime();
        // support v1 codepath and v2 entrypoint like 'foo', '', 'index.hello'
        let userFunction = funcname
            ? require(modulepath)[funcname]
            : require(modulepath);
        let elapsed = process.hrtime(startTime);
        console.log(
            `user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`
        );
        return userFunction;
    } catch (e) {
        console.error(`user code load error: ${e}`);
        return e;
    }
};

// Request logger
if (process.env["KYMA_INTERNAL_LOGGER_ENABLED"]) {
    app.use(morgan("combined"));
}


app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: `${bodySizeLimit}mb`, strict: false  }))
app.use(bodyParser.text({ type: ['text/*'], limit: `${bodySizeLimit}mb`  }))
app.use(bodyParser.urlencoded({ limit: `${bodySizeLimit}mb`, extended: true }));
app.use(bodyParser.raw({limit: `${bodySizeLimit}mb`, type: () => true}))

app.use(helper.handleTimeOut);

app.get("/healthz", (req, res) => {
    res.status(200).send("OK")
})

app.get("/metrics", (req, res) => {
    getMetrics(req, res)
})

app.get('/favicon.ico', (req, res) => res.status(204));

// Generic route -- all http requests go to the user function.
app.all("*", (req, res, next) => {


    res.header('Access-Control-Allow-Origin', '*');
    if (req.method === 'OPTIONS') {
        // CORS preflight support (Allow any method or header requested)
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
    } else {
    
        callsTotalCounter.add(1)
        const startTime = new Date().getTime()

        if (!userFunction) {
            failuresTotalCounter.add(1)
            res.status(500).send("User function not loaded");
            return;
        }

        const event = ce.buildEvent(req, res, tracer);

        const context = {
            'function-name': functionName,
            'runtime': process.env.FUNC_RUNTIME,
            'namespace': serviceNamespace
        };

        const sendResponse = (body, status, headers) => {
            if (res.writableEnded) return;
            if (headers) {
                for (let name of Object.keys(headers)) {
                    res.set(name, headers[name]);
                }
            }
            if(body){
                if(status){
                    res.status(status);
                } 
                switch (typeof body) {
                    case 'object':
                        res.json(body); // includes res.end(), null also handled
                        break;
                    case 'undefined':
                        res.end();
                        break;
                    default:
                        res.end(body);
                }
                // res.send(body);
            } else if(status){
                res.sendStatus(status);
            } else {
                res.end();
            }
        };

        const span = startNewSpan('userFunction', tracer);

        try {
            // Execute the user function
            const out = userFunction(event, context, sendResponse);
            if (out && helper.isPromise(out)) {
                out.then(result => {
                    sendResponse(result)
                })
                .catch((err) => {
                    helper.handleError(err, span, sendResponse)
                    failuresTotalCounter.add(1);
                })
                .finally(()=>{
                    span.end();
                })
            } else {
                sendResponse(out);
            }
        } catch (err) {
            helper.handleError(err, span, sendResponse)
            failuresTotalCounter.add(1);
        } finally {
            span.end();
        }

        const endTime = new Date().getTime()
        const executionTime = endTime - startTime;
        durationHistogram.record(executionTime);
    }
});


const server = app.listen(funcPort);

helper.configureGracefulShutdown(server);


const fn = loadFunction("./function/handler", "");
if (helper.isFunction(fn.main)) {
    userFunction = fn.main
} else {
    console.error("Content loaded is not a function", fn)
}