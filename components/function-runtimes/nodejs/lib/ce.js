'use strict';

const axios = require('axios');
const { HTTP, CloudEvent } = require('cloudevents');
const charset = 'utf-8'

const publishProxyAddress = (process.env.PUBLISHER_PROXY_ADDRESS);

module.exports = {
    buildEvent
};

function buildEvent(req, res, tracer) {

    let event = {
        tracer,
        'extensions': { request: req, response: res },
        setResponseHeader: (key, value) => setResponseHeader(res, key, value),
        setResponseContentType: (type) => setResponseContentType(res, type),
        setResponseStatus: (status) => setResponseStatus(res, status),
        //deprecated
        publishCloudEvent: (data) => publishCloudEvent(data),
        //deprecated
        buildResponseCloudEvent: (id, type, data) => buildResponseCloudEvent(req, id, type, data),
        emitCloudEvent: (type, source, data, optionalCloudEventAttributes) => emitCloudEvent(type, source, data, optionalCloudEventAttributes),
    };

    if(req.body){
        if (!req.is('multipart/*')) {
            if(isCloudEvent(req)) {
                event = Object.assign(event,buildCloudEventAttributes(req));    
            } else {
                event = Object.assign(event,{'data':req.body});
            }
        }
    }
    return event;
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
    console.warn("publishCloudEvent is deprecated. Use emitCloudEvent")
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
    console.warn("buildResponseCloudEvent is deprecated. Use emitCloudEvent")
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

function isCloudEvent(req) {
    return req.is('application/cloudevents+json') || hasCeHeaders(req);
}


function hasCeHeaders(req) {
    return req.get('ce-type') && req.get('ce-source');
}

function buildCloudEventAttributes(req) {
    const receivedEvent = HTTP.toEvent({ headers: req.headers, body: req.body });  
    return {
        'ce-type': receivedEvent.type,
        'ce-source': receivedEvent.source,
        'ce-eventtypeversion': receivedEvent.eventtypeversion,
        'ce-specversion': receivedEvent.specversion,
        'ce-id': receivedEvent.id,
        'ce-time': receivedEvent.time,
        'ce-datacontenttype': receivedEvent.datacontenttype,
        'data': receivedEvent.data
    };
}

function emitCloudEvent(type, source, data, optionalCloudEventAttributes) {

    let optionalCloudEventAttributesInput = optionalCloudEventAttributes
    if(!optionalCloudEventAttributesInput){
        optionalCloudEventAttributesInput = {}
    }

    let cloudEventInput = {
        type,
        source,
        data,
    }

    if(!optionalCloudEventAttributesInput.datacontenttype){
        optionalCloudEventAttributesInput.datacontenttype = resolvedatatype(data);
    }
    
    cloudEventInput = Object.assign(cloudEventInput, optionalCloudEventAttributesInput)
    const ce = new CloudEvent(cloudEventInput);
    const message = HTTP.structured(ce)

    return axios({
        method: "post",
        baseURL: publishProxyAddress,
        headers: message.headers,
        data: message.body,
    });
}