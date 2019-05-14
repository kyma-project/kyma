import http from "k6/http"
import { check, sleep } from "k6";

export let options = {
  vus: 10,
  duration: "1m",
  rps: 1000,
  tags: {
    "testName": "size-l",
    "component": "serverless",
    "revision": `${__ENV.REVISION}`
  }
}

export let configuration = {
  url: `https://size-l.${__ENV.CLUSTER_DOMAIN_NAME}`,
}

export default function () {
  let res = http.get(configuration.url);

  check(res, {
    "status was 200": (r) => r.status == 200,
    "transaction time OK": (r) => r.timings.duration < 200
  });
};