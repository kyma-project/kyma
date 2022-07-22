'use strict';

const axios = require('axios');

const publishProxyAddress = (process.env.PUBLISHER_PROXY_ADDRESS);

module.exports = {
    buildEvent
};

function buildEvent(req, res, tracer) {
    let data = req.body;
    if (!req.is('multipart/*') && req.body.length > 0) {
        if (req.is('application/json')) {
            data = JSON.parse(req.body.toString('utf-8'))
        } else {
            data = req.body.toString('utf-8')
        }
    }

    return Object.assign( 
        buildCeHeaders(req), {
        data,
        'extensions': { request: req, response: res },
        setResponseHeader: (key, value) => setResponseHeader(res, key, value),
        setResponseContentType: (type) => setResponseContentType(res, type),
        setResponseStatus: (status) => setResponseStatus(res, status),
        publishCloudEvent: (data) => publishCloudEvent(data),
        buildResponseCloudEvent: (id, type, data) => buildResponseCloudEvent(req, id, type, data),
        tracer
    });
}

function setResponseHeader(res, key, value) {
    res.set(key, value);
}

function setResponseContentType(res, type) {
    res.type(type);
}

function setResponseStatus(res, status) {
    res.status(status);
}

function publishCloudEvent(data) {
    return axios({
        method: "post",
        baseURL: publishProxyAddress,
        headers: {
            "Content-Type": "application/cloudevents+json"
        },
        data: data,
    });
}

function resolvedatatype(data){
    switch(typeof data) {
        case 'object':
            return 'application/json'
        case 'string':
            return 'text/plain'
    }
}

function buildResponseCloudEvent(req, id, type, data) {
    return {
        'type': type,
        'source': req.get('ce-source'),
        'eventtypeversion': req.get('ce-eventtypeversion'),
        'specversion': req.get('ce-specversion'),
        'id': id,
        'data': data,
        'datacontenttype': resolvedatatype(data),
    };
}

function buildCeHeaders(req) {
    return {
        'ce-type': req.get('ce-type'),
        'ce-source': req.get('ce-source'),
        'ce-eventtypeversion': req.get('ce-eventtypeversion'),
        'ce-specversion': req.get('ce-specversion'),
        'ce-id': req.get('ce-id'),
        'ce-time': req.get('ce-time'),
    };
}
