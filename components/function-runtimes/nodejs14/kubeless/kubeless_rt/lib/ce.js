'use strict';

const axios = require('axios');

const publishProxyAddress = (process.env.PUBLISHER_PROXY_ADDRESS); // "http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish"

module.exports = {
    buildEvent
};

function buildEvent(req, res) {
    let data = req.body;
    if (!req.is('multipart/*') && req.body.length > 0) {
        if (req.is('application/json')) {
            data = JSON.parse(req.body.toString('utf-8'))
        } else {
            data = req.body.toString('utf-8')
        }
    }

    return Object.assign( getCeHeaders(req), {
        data,
        'extensions': { request: req, response: res },
        sendRespondEvent: (event) => sendRespondEvent(req, event),
        publishCloudEvent: (data) => publishCloudEvent(data),
        buildCloudEvent: (id, type, data) => buildCloudEvent(req, id, type, data)
    });
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

function sendRespondEvent(ce, data) {
    let response = ce.res;
    for (let [key, val] of Object.entries(getCeHeaders(ce))) {
        response.set(key, val);
    }

    response.send(data);
}

function buildCloudEvent(req, id, type, data) {
    return {
        'type': type,
        'source': req.get('ce-source'),
        'eventtypeversion': req.get('ce-eventtypeversion'),
        'specversion': req.get('ce-specversion'),
        'id': id,
        'data': data,
    };
}

function getCeHeaders(req) {
    return {
        'ce-type': req.get('ce-type'),
        'ce-source': req.get('ce-source'),
        'ce-eventtypeversion': req.get('ce-eventtypeversion'),
        'ce-specversion': req.get('ce-specversion'),
        'ce-id': req.get('ce-id'),
        'ce-time': req.get('ce-time'),
    };
}