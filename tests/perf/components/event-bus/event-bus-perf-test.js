import http from "k6/http";
import {
    check
} from "k6";
import {
    Counter,
    Trend
} from "k6/metrics";

var eventPublishTotalRequests = new Counter("event_publish_requests"),
    eventPublishedSuccessfully = new Counter("event_publish_success"),
    eventPublishedFailed = new Counter("event_published_failed"),
    eventPublishedResponseTime = new Trend("event_publish_response_time", true)


export let options = {
    rps: 0,
    tags: {
        component: "event-bus",
        "testName": "event-publish-service",
        revision: `${__ENV.REVISION}`
    },

    /**
     * Bursty traffic,
     * Measure the worst case of bursty traffic, when bursting results in scaling
     * Bursts: 0 -> N, and N -> 2N
     * where, N = {1, 100, 1000}
     * 
     * Note: this is the desired KPI, which is currently not possible due 
     * to https://github.com/kyma-project/kyma/issues/4370
     */

    /**
     * Numbers have to be brought down as the Framework kills the test
     * w.r.t https://github.com/kyma-project/kyma/issues/4370 
     */
    stages: [
        //Expected
        //     { duration: "1m", target: 1 },
        //     { duration: "1m", target: 2 },
        //     { duration: "1m", target: 10 },
        //     { duration: "1m", target: 20 },
        //     { duration: "1m", target: 100 },
        //     { duration: "1m", target: 200 }
        {
            duration: "90s",
            target: 2 ** 3
        },
        {
            duration: "90s",
            target: 2 ** 4
        },
        {
            duration: "90s",
            target: 2 ** 5
        }
    ],
    noConnectionReuse: true,
    tlsAuth: [{
        domains: [`gateway.${__ENV.CLUSTER_DOMAIN_NAME}`],
        cert: open(`${__ENV.APP_CONNECTOR_CERT_DIR}/generated.crt`),
        key: open(`${__ENV.APP_CONNECTOR_CERT_DIR}/generated.key`)
    }]
};

export let event_publish_configuration = {
    params: {
        headers: {
            "Content-Type": "application/json"
        }
    },
    url: `https://gateway.${__ENV.CLUSTER_DOMAIN_NAME}/perf-app/v1/events`,
    payload: JSON.stringify({
        "source-id": "perf-app",
        "event-type": "hello",
        "event-type-version": "v1",
        "event-time": "2018-11-02T22:08:41+00:00",
        "data": "some-event"
    })
}

export default function () {
    var eventPublishResponse = http.post(event_publish_configuration.url, event_publish_configuration.payload, event_publish_configuration.params);
    eventPublishedResponseTime.add(eventPublishResponse.timings.duration);

    check(eventPublishResponse, {
        "response code was 200": (res) => res.status == 200,
        "event publish status is 'published'.": (res) => {
            var a = JSON.parse(res.body)
            if (a && a.status && (a.status == "published")) {
                return true
            } else {
                return false;
            }
        }
    }) ? eventPublishedSuccessfully.add(1) : eventPublishedFailed.add(1);

    eventPublishTotalRequests.add(1) //Increment number of events
};
