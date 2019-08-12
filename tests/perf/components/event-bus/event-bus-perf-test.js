import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

var eventPublishTotalRequests = new Counter("event_publish_requests");
var eventPublishedSuccessfully = new Counter("event_publish_success");
var eventPublishedFailed = new Counter("event_published_failed");

/*
    Subscription relevant metrics, TO be enabled once subscriber seems to be thread safe
*/
// var subscriptionResultVerified = new Counter("verified_subscription_results");
// var subscriptionResultFailed = new Counter("failed_subscription_results");

let eventPublishUrl = "http://event-publish-service.kyma-system:8080/v1/events";
   // eventSubscriberUrl = "http://event-subscription-service.event-bus-perf-test:9000/v1/results";

var params =  { headers: { "Content-Type": "application/json" } },
eventPublishPayload = `{"source-id": "external-application", "event-type": "hello", `+
                        `"event-type-version": "v1", "event-time": "2018-11-02T22:08:41+00:00", "data": "some-event"}`;

export let options = {
  rps: 10,
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
  stages: [
    { duration: "1m", target: 1 },
    { duration: "1m", target: 2 },
    { duration: "1m", target: 10 },
    { duration: "1m", target: 20 },
    { duration: "1m", target: 100 },
    { duration: "1m", target: 200 }
  ],
  noConnectionReuse: true
};

export default function () {
    var eventPublishResponse = http.post(eventPublishUrl, eventPublishPayload, params);
    check(eventPublishResponse, {
        "response code was 200": (res) => res.status == 200,
        "event publish status is 'published'.": (res) => {
            var a = JSON.parse(res.body)
            if(a && a.status && (a.status == "published")) {
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

