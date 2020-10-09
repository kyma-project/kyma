'use strict';

const vm = require('vm');
const path = require('path');
const Module = require('module');

const bodyParser = require('body-parser');
const client = require('prom-client');
const express = require('express');
const helper = require('./lib/helper');
const morgan = require('morgan');

const bodySizeLimit = Number(process.env.REQ_MB_LIMIT || '1');

const app = express();
app.use(morgan('combined'));
const bodParserOptions = {
    type: req => !req.is('multipart/*'),
    limit: `${bodySizeLimit}mb`,
};
app.use(bodyParser.raw(bodParserOptions));
app.use(bodyParser.json({ limit: `${bodySizeLimit}mb` }));
app.use(bodyParser.urlencoded({ limit: `${bodySizeLimit}mb`, extended: true }));

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;
const timeout = Number(process.env.FUNC_TIMEOUT || '180');
const funcPort = Number(process.env.FUNC_PORT || '8080');

const modKubeless = require.main.filename;
const modRootPath = path.join(modKubeless, '..', '..', 'kubeless');

const modPath = path.join(modRootPath, `${modName}.js`);
const libPath = path.join(modRootPath, 'node_modules');
const pkgPath = path.join(modRootPath, 'package.json');
const libDeps = helper.readDependencies(pkgPath);

const { timeHistogram, callsCounter, errorsCounter } = helper.prepareStatistics('method', client);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);

const context = {
    'function-name': funcHandler,
    'timeout' : timeout,
    'runtime': process.env.FUNC_RUNTIME,
    'memory-limit': process.env.FUNC_MEMORY_LIMIT
};

const script = new vm.Script('\nrequire(\'kubeless\')(require(\''+ modPath +'\'));\n', {
    filename: modPath,
    displayErrors: true,
});

function modRequire(p, req, res, end) {
    if (p === 'kubeless')
        return (handler) => modExecute(handler, req, res, end);
    else if (libDeps.includes(p))
        return require(path.join(libPath, p));
    else if (p.indexOf('./') === 0)
        return require(path.join(path.dirname(modPath), p));
    else
        return require(p);
}

function modExecute(handler, req, res, end) {
    let func = null;
    switch (typeof handler) {
        case 'function':
            func = handler;
            break;
        case 'object':
            if (handler) func = handler[funcHandler];
            break;
    }
    if (func === null)
        throw new Error(`Unable to load ${handler}`);

    try {
        let data = req.body;
        if (!req.is('multipart/*') && req.body.length > 0) {
            if (req.is('application/json')) {
                data = JSON.parse(req.body.toString('utf-8'))
            } else {
                data = req.body.toString('utf-8')
            }
        }
        const event = {
            'ce-type': req.get('ce-type'),
            'ce-source': req.get('ce-source'),
            'ce-eventtypeversion': req.get('ce-eventtypeversion'),
            'ce-specversion': req.get('ce-specversion'),
            'ce-id': req.get('ce-id'),
            'ce-time': req.get('ce-time'),
            data,
            'extensions': { request: req, response: res },
        };
        Promise.resolve(func(event, context))
        // Finalize
            .then(rval => modFinalize(rval, res, end))
            // Catch asynchronous errors
            .catch(err => handleError(err, res, funcLabel(req), end))
        ;
    } catch (err) {
        // Catch synchronous errors
        handleError(err, res, funcLabel(req), end);
    }
}

function modFinalize(result, res, end) {
    if (!res.finished) switch(typeof result) {
        case 'string':
            res.end(result);
            break;
        case 'object':
            res.json(result); // includes res.end(), null also handled
            break;
        case 'undefined':
            res.end();
            break;
        default:
            res.end(JSON.stringify(result));
    }
    end();
}

function handleError(err, res, label, end) {
    errorsCounter.labels(label).inc();
    res.status(500).send('Internal Server Error');
    console.error(`Function failed to execute: ${err.stack}`);
    end();
}

function funcLabel(req) {
    return modName + '-' + req.method;
}

app.all('*', (req, res) => {
    res.header('Access-Control-Allow-Origin', '*');
    if (req.method === 'OPTIONS') {
        // CORS preflight support (Allow any method or header requested)
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
    } else {
        const label = funcLabel(req);
        const end = timeHistogram.labels(label).startTimer();
        callsCounter.labels(label).inc();

        const sandbox = Object.assign({}, global, {
            __filename: modPath,
            __dirname: modRootPath,
            module: new Module(modPath, null),
            require: (p) => modRequire(p, req, res, end),
        });

        try {
            script.runInNewContext(sandbox, { timeout : timeout * 1000 });
        } catch (err) {
            if (err.toString().match('Error: Script execution timed out')) {
                res.status(408).send(err);
                // We cannot stop the spawned process (https://github.com/nodejs/node/issues/3020)
                // we need to abruptly stop this process
                console.error('CRITICAL: Unable to stop spawned process. Exiting');
                process.exit(1);
            } else {
                handleError(err, res, funcLabel, end);
            }
        }
    }
});

const server = app.listen(funcPort);
helper.configureGracefulShutdown(server);
