import http from 'k6/http';
import { Trend } from "k6/metrics";
import { check, sleep } from "k6";

export let options = {
    vus: 10,
    stages: [
        { duration: "2m", target: 10 },
        { duration: "1m", target: 25 },
        { duration: "1m", target: 50 },
        { duration: "2m", target: 100 },
        { duration: "1m", target: 50 },
        { duration: "1m", target: 25 },
        { duration: "2m", target: 10 },
    ],
    rps: 1000,
    tags: {
        "testName": "istio_load_test",
        "component": "httpbin",
        "revision": `${__ENV.REVISION}`
    },
    conf: {
        workerCount: `${__ENV.WORKLOAD_SIZE}`,
        domain: `${__ENV.CLUSTER_DOMAIN_NAME}`
    }
}

var istioTrend = new Trend("istio_request_time", true);

export default function() {
    var randomnumber = Math.floor(Math.random() * (options.conf.workerCount)) + 1;
    const response = http.get(`https://httpbin-${randomnumber}.${options.conf.domain}/cookies`);

    istioTrend.add(response.timings.duration);

    check(response, {
        "status was 200": (r) => r.status == 200,
        "transaction time < 1000 ms": (r) => r.timings.duration < 1000
    });
    sleep(1);
}