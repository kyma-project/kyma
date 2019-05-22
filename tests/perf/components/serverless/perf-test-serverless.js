import http from "k6/http"
import { check } from "k6";

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

export let configuration =
  [
    {
      url: `https://size-s.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-s'
        }
      }
    },
    {
      url: `https://size-m.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-m'
        }
      }
    },
    {
      url: `https://size-l.${__ENV.CLUSTER_DOMAIN_NAME}`,
      tags: {
        tags: {
          'testName': 'size-l'
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

// each virtual user runs this function in a loop
export default function () {
  // call all lambda functions
  configuration.forEach(function (element) {
    let res = http.get(element.url, element.tags);
    
    // TODO: report custom metric where delay is substracted
    check(res, {
      "status was 200": (r) => r.status == 200,
    }, element.tags);
  });
};