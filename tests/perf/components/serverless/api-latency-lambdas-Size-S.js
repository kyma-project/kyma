import http from "k6/http"
import { check, sleep } from "k6";

export let options = {
  vus: 10,
  rps: 1000,
  tags: {
    "component": "serverless",
    "revision": `${__ENV.REVISION}`
  },
  // ramp up #virtual users (VU) over time to get maximum througput
  stages: [
    { duration: "90s", target: 10 },
    { duration: "90s", target: 100 },
    { duration: "90s", target: 1000 },
    // { duration: "30s", target: 10000 },
  ],
}

export let configuration =
  [
    {
      url: `https://size-s.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-sc'
        }
      }
    },
    {
      url: `https://size-m.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-mc'
        }
      }
    },
    {
      url: `https://size-l.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-lc'
        }
      }
    },
    // {
    //   url: `https://size-xl.${__ENV.CLUSTER_DOMAIN_NAME}`,
    //   tags: {
    //     tags: {
    //       'testName': 'size-xlc'
    //     }
    //   }
    // }
  ]

export default function () {
  configuration.forEach(function (element) {
    let res = http.get(element.url, element.tags);

    check(res, {
      "status was 200": (r) => r.status == 200,
      "transaction time OK": (r) => r.timings.duration < 200
    }, element.tags);
  });
};