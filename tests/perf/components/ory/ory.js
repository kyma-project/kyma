import http from 'k6/http';
import { check, sleep } from "k6";
import encoding from "k6/encoding";

export let options = {
    vus: 10,
    duration: "10s",
    rps: 10,
    tags: {
        "testName": "http_db_service_10vu_60s_1000",
        "component": "http-db-service",
        "revision": `${__ENV.REVISION}`
    },
    conf: {
        clientId: `${__ENV.CLIENT_ID}`,
        clientSecret: `${__ENV.CLIENT_SECRET}`,
        domain: `${__ENV.CLUSTER_DOMAIN}`
    }
};


// export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
// -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=read"
// curl -ik -X GET https://httpbin.$DOMAIN/headers -H "Authorization: Bearer $ACCESS_TOKEN_READ"

export function setup() {
    const credentials = encoding.b64encode(`${options.conf.clientId}:${options.conf.clientSecret}`);

    console.log(credentials);

    let url = `https://oauth2.${options.conf.domain}/oauth2/token`;
    let payload = { "grant_type": "client_credentials", "scope": "read", client_id: options.conf.clientId };
    let params =  { headers: { "Authorization": `Basic ${credentials}` }};

    let accessToken = http.post(url, payload, params);
    console.log("token: " + accessToken.body);

    return accessToken.body
}

export default function(data) {
    let token = JSON.parse(data).access_token;
    let url = `https://httpbin1.${options.conf.domain}/headers`;
    let params =  { headers: { "Authorization": `Bearer ${token}` }};

    const response = http.get(url, params);

    check(response, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 500
    });
    sleep(1);
}

