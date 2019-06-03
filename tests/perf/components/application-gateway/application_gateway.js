import http from "k6/http"
import { check, sleep } from "k6";

export let options = {
    vus: 16,
    duration: "3m",
    rps: 1000,
    tags: {
        "testName": "send_event_gateway_16vu_180s_1000",
        "component": "application-gateway",
        "revision": `${__ENV.REVISION}`
    },
    tlsAuth: [
        {
            domains: [`gateway.${__ENV.CLUSTER_DOMAIN_NAME}`],
            cert: open(`${__ENV.APP_CONNECTOR_CERT_DIR}/generated.crt`),
            key: open(`${__ENV.APP_CONNECTOR_CERT_DIR}/generated.key`)
        }
    ]
}

export let configuration = {
    params: { headers: { "Content-Type": "application/json" } },
    url: `https://gateway.${__ENV.CLUSTER_DOMAIN_NAME}/perf-app/v1/events`,
    payload: JSON.stringify(
        {
            "event-type": "petCreated",
            "event-type-version": "v1",
            "event-time": "2018-11-02T22:08:41+00:00",
            "data": {
                "pet": {
                    "id": "4caad296-e0c5-491e-98ac-0ed118f9474e"
                }
            }
        })
}

export default function () {
    let res = http.post(configuration.url, configuration.payload, configuration.params);

    check(res, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 200
    });
    sleep(1);
};