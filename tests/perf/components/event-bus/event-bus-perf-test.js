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
    eventPublishedResponseTime = new Trend("event_publish_response_time")

/*
    Subscription relevant metrics, TO be enabled once subscriber seems to be thread safe
*/
// var subscriptionResultVerified = new Counter("verified_subscription_results");
// var subscriptionResultFailed = new Counter("failed_subscription_results");

// let eventSubscriberUrl = "http://event-subscription-service.event-bus-perf-test:9000/v1/results";


export let options = {
    rps: 0,
    tags: {
        component: "event-bus",
        revision: `${__ENV.REVISION}`
    },

    /**
     * Bursty traffic,
     * Measure the worst case of bursty traffic, when bursting results in scaling
     * Bursts: 0 -> N, and N -> 2N
     * where, N = {1, 100, 1000}
     */

    /**
     * Numbers have to be brought down as the Framework kills the test
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

    // var subscriptionResultResponse = http.get(eventSubscriberUrl);
    // check(subscriptionResultResponse, {
    //     "response code was 200": (res) => res.status == 200
    // }) ? subscriptionResultVerified.add(1) : subscriptionResultFailed.add(1);

    eventPublishTotalRequests.add(1) //Increment number of events
};