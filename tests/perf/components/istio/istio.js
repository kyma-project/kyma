import http from 'k6/http';
import { check, group } from "k6";
import encoding from "k6/encoding";
import { Trend } from "k6/metrics";

export let options = {
    vus: 1,
    duration: "1m",
    rps: 10,
    tags: {
        "testName": "istio_10vu_60s_100",
        "component": "istio",
        "revision": `${__ENV.REVISION}`
    },
    conf: {
        domain: `${__ENV.CLUSTER_DOMAIN_NAME}`
    }
};

export function setup() {
    // console.log("setup called");
}

export default function(data) {
    // console.log("default called");

    group("get cookies", function() {
        let url = `https://httpbin.${options.conf.domain}/cookies`;
        const response = http.get(url);

        //Custom metrics
        // oauth2Trend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "true"});
    });
}