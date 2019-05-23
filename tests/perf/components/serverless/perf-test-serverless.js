import http from "k6/http"
import { check } from "k6"
import { Trend } from "k6/metrics"

var httpReqDurationNoFunc = new Trend("http_req_duration_no_func", true)
var httpReqWaitingNoFunc = new Trend("http_req_waiting_no_func", true)

export let options = {
  // unlimited
  rps: 0,
  tags: {
    "component": "serverless",
    "revision": `${__ENV.REVISION}`
  },
  // ramp up #virtual users (VU) over time to get maximum throughput
  stages: [
    { duration: "90s", target: 10 },
    { duration: "90s", target: 100 },
    { duration: "90s", target: 1000 },
  ],
}

export let functionNames =
  ['size-s',
    'size-l',
    'size-m',
  ]

// each virtual user runs this function in a loop
export default function () {
  let funcDelay = parseInt(`${__ENV.FUNC_DELAY}`)

  // call all lambda functions
  functionNames.forEach(function (functionName) {
    let url = `https://${functionName}.${__ENV.CLUSTER_DOMAIN_NAME}`
    let tags = {
      'testName': functionName,
    }
    let res = http.get(url, { 'tags': tags })

    // report custom trend
    // same as http_req_duration but without function execution time
    httpReqDurationNoFunc.add(res.timings.duration - funcDelay);
    // same as http_req_waiting but without function execution time
    httpReqWaitingNoFunc.add(res.timings.waiting - funcDelay)

    check(res, {
      "status was 200": (r) => r.status == 200,
    }, tags);
  });
};