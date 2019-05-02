# Kyma Performance Test Guidelines

## Overview
Kyma uses [K6](https://docs.k6.io/docs) for performance and load testing, K6 is a developer oriented well documented load testing tool, easy to 
integrate into Kyma workflow and automation flow.

### Running K6
To execute a very simple k6 test from the command line, you can use this:
```bash
k6 run github.com/kyma-project/kyma/tests/perf/examples/http_basic.js
```

With command above, k6 will fetch the http_het.js file from Github and only fires 1 virtual user which get execute 
the script once.

K6 can also execute local files.
```bash
k6 run script.js
```

## K6 with Kyma

Directory ```tests/perf``` contains all performance tests source code.
A Kyma performance test is a K6 test script with or without prerequisites like custom component deployments or configurations.
A Kyma performance test will runs against [Kyma load test cluster](https://github.com/kyma-project/test-infra).

Each subdirectory in the ```tests/perf``` directory defines source code for one test suite and focusing on one component, 
the subdirectory ```prerequisites``` will contains **yaml** files of custom component deployments 
(like custom configuration or custom scenario deployments) if necessary.

Prerequisites directory content will be deployed after load test cluster deployment and before test execution.

### Implementing Kyma performance test

This section will document Kyma specific k6 test implementation, for detailed information about k6 test framework you can 
read from [original documentation](https://docs.k6.io/docs)

Kyma k6 executor has some pre-defined environment variable and tags to provide some additional information about 
current execution and target test cluster.

More about K6 tag please read from [here](https://docs.k6.io/docs/tags-and-groups).

Environment variables
- **CLUSTER_DOMAIN_NAME**, is the domain name of the target Kyma load test cluster
- TBD

Test execution tags 
- **testName**, is the name of test which every test should provide in test code implementation. 
This information will be used later on [grafana](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1) to filter test results
- TBD

An example k6 test testing Kyma Gateway, with predefined tag name ```testName```

```javascript
import http from "k6/http"
import {check, sleep} from "k6";

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    tags: {
        "testName": "gateway_event"
    }
}

export let configuration = {
    params: {headers: {"Content-Type": "application/json"}},
    url: `https://gateway.${__ENV.CLUSTER_DOMAIN_NAME}/lszymik/v1/events`,
    payload: JSON.stringify({
        "event-type": "petCreated",
        "event-type-version": "v1",
        "event-time": "2018-11-02T22:08:41+00:00",
        "data": {"pet": {"id": "4caad296-e0c5-491e-98ac-0ed118f9474e"}}
    })
}

export default function () {
    let res = http.post(configuration.url, configuration.payload, configuration.params);

    check(res, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 200
    });
    sleep(1);
}
```

Example test above will execute a load test against Kyma gateway on a cluster deployed on **CLUSTER_DOMAIN_NAME** 
with 10 virtual users, 1 minute long and 1000 request per second across all virtual users.

Test logic should be implemented in a function defined as default, more about test execution lifecycle please read [here](https://docs.k6.io/docs/test-life-cycle).

Variable ```options``` defines execution behavior of test implementation. With following options

- ```vus``` defines amount of virtual users.
- ```duration``` defines test execution duration.
- ```rps``` defines request per second across all virtual users
- ```tags``` defines custom tags to mark test execution

More about available options and test execution behavior please read [here](https://docs.k6.io/docs/options).

Result of test execution will be stored in a Influx-DB along with Grafana on Kyma Load Generator Cluster and can be accessed from [here](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1)

### Run test locally