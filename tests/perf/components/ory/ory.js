import http from 'k6/http';
import { check, group, fail, sleep } from "k6";
import encoding from "k6/encoding";
import { Trend } from "k6/metrics";

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    setupTimeout: "20s",
    tags: {
        "testName": "ory_10vu_60s_100",
        "component": "ory",
        "revision": `${__ENV.REVISION}`
    },
    conf: {
        clientId: `${__ENV.CLIENT_ID}`,
        clientSecret: `${__ENV.CLIENT_SECRET}`,
        domain: `${__ENV.CLUSTER_DOMAIN_NAME}`
    }
};

export function setup() {
    const credentials = encoding.b64encode(`${options.conf.clientId}:${options.conf.clientSecret}`);

    let url = `https://oauth2.${options.conf.domain}/oauth2/token`;
    let payload = { "grant_type": "client_credentials", "scope": "read", client_id: options.conf.clientId };
    let params =  { headers: { "Authorization": `Basic ${credentials}` }};

    let hydraResponse = retry(url, payload, params, 5);
    if (hydraResponse.error_code != 0) {
        fail("Unable to fetch token. Error code: " + hydraResponse.error_code + ", hydra response: " + hydraResponse.body);
    }

    let token = JSON.parse(hydraResponse.body).access_token;
    return token
}

let oauth2Trend = new Trend("oauth2_req_time", true);
let oauth2WithTokenTrend = new Trend("oauth2_with_token_req_time", true);
let oauth2IDTokenMutatorTrend = new Trend("oauth2_id_token_mutator_req_time", true);
let oauth2HeaderMutatorTrend = new Trend("oauth2_header_mutator_req_time", true);
let noopTrend = new Trend("noop_req_time", true);
let allowTrend = new Trend("allow_req_time", true);

export default function(token) {
    let params = {headers: {"Authorization": `Bearer ${token}`}};
    group("get oauth2 secured service", function() {
        let url = `https://httpbin1.${options.conf.domain}/headers`;
        const response = http.get(url, params);

        //Custom metrics
        oauth2Trend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "true"});
    });

    group("get token and oauth2 secured service", function() {
        let url = `https://httpbin5.${options.conf.domain}/headers`;
        const response = http.get(url, params);

        //Custom metrics
        oauth2WithTokenTrend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "true"});
    });

    group("get oauth2 secured service with id token mutator", function() {
        let url = `https://httpbin2.${options.conf.domain}/headers`;
        const response = http.get(url, params);

        //Custom metrics
        oauth2IDTokenMutatorTrend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "true"});
    });

    group("get oauth2 secured service with header mutator", function() {
        let url = `https://httpbin3.${options.conf.domain}/headers`;
        const response = http.get(url, params);

        //Custom metrics
        oauth2HeaderMutatorTrend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "true"});
    });

    group("get not secured service with noop", function() {
        let url = `https://httpbin.${options.conf.domain}/headers`;
        const response = http.get(url);

        //Custom metrics
        noopTrend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "false"});
    });

    group("get not secured service with allow", function() {
        let url = `https://httpbin4.${options.conf.domain}/headers`;
        const response = http.get(url);

        //Custom metrics
        allowTrend.add(response.timings.duration);

        //Check
        check(response, {
            "status was 200": (r) => r.status == 200,
            "transaction time < 1000 ms": (r) => r.timings.duration < 1000
        }, {secured: "false"});
    });
}

function retry(url, payload, params, n) {
    let res;
    for (var retries = n; retries > 0; retries--) {
        res = http.post(url, payload, params);
        console.log(retries);
        if (res.error_code == 0) {
            return res;
        }
        sleep(3)
    }
    return res;
}