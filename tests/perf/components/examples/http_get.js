import http from 'k6/http';

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    tags: {
        "testName": "http_basic_10vu_60s_1000",
        "component": "http-db-service",
        "revision": `${__ENV.REVISION}`
    }
}

export default function() {
    const response = http.get(`https://http-db-service.${__ENV.CLUSTER_DOMAIN_NAME}/`);
}