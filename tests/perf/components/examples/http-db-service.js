import http from 'k6/http';
import { check, sleep } from "k6";

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    tags: {
        "testName": "http_db_service_10vu_60s_1000",
        "component": "http-db-service",
        "revision": `${__ENV.REVISION}`
    }
}

export default function() {
    const response = http.get(`https://http-db-service.${__ENV.CLUSTER_DOMAIN_NAME}/`);

    check(response, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 200
    });
    sleep(1);
}