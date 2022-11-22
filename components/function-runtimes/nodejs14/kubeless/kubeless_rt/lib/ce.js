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
        emitCloudEvent: (type, source, data, eventtypeversion) => emitCloudEvent(type, source, data, eventtypeversion),
    };

    if (req.body){
        let stringifiedReqestBody = req.body.toString(charset);
        if (!req.is('multipart/*') && req.body.length > 0) {
            if(isCloudEvent(req)) {
                event = Object.assign(event,buildCloudEventAttributes(req, JSON.parse(stringifiedReqestBody)));    
            } else if (isOfJsonContentType(req)) {
                event = Object.assign(event,{'data':JSON.parse(stringifiedReqestBody)});
            } else {
                event = Object.assign(event,{'data':stringifiedReqestBody});
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

function isOfJsonContentType(req) {
    return req.is('application/json');
}

function hasCeHeaders(req) {
    return req.get('ce-type') && req.get('ce-source');
}

function buildCloudEventAttributes(req, data) {
    const receivedEvent = HTTP.toEvent({ headers: req.headers, body: data });  
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
    return axios({
        method: "post",
        baseURL: publishProxyAddress,
        headers: {
            "Content-Type": "application/cloudevents+json"
        },
        data: ce.toJSON(),
    });
}
