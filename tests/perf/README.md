# Performance Tests

## Overview

Kyma uses [K6](https://docs.k6.io) as a tool for performance and load testing. One of its benefits is that it is easy to integrate it with the Kyma workflow and automation flow.

### Folder structure

The `tests/perf` directory contains the source code for the performance test script.
A Kyma performance test is a K6 test script that can contain prerequisites such as a custom test component or scenario deployments. Every test runs against the [Kyma load test cluster](https://github.com/kyma-project/test-infra).

Each subdirectory in the `tests/perf/components` directory defines source code for one test suite and refers to one component or area.
The `prerequisites` subdirectory contains shell script with name **`setup.sh`** and `yaml` files of test component deployments and, if required, shell scripts such as custom configuration or custom scenario deployments.
There should be one `prerequisites` subdirectory for each component. The name of this subdirectory should be the same as the name of a given subdirectory with scripts under the `components` subdirectory.
If test need some custom scenario deployment, you should provide a shell script with name **`setup.sh`** to bootstrap custom scenario deployment.

See an exemplary directory structure for **application-gateway** and **event-bus** components:
```
tests
|  
+--- perf
     |   
     |--- components
     |    |
     |    |--- application-gateway // A folder with test scripts for the Application Gateway
     |    |
     |    +--- event-bus           // A folder with test scripts for the Event Bus
     |      
     +--- prerequisites
          |
          |--- application-gateway // A folder with shell scripts and deployment files for the Application Gateway
          |
          +--- event-bus           // A folder with shell scripts and deployment files for the Event Bus

```
The content of the `prerequisites` subdirectory is deployed after load test cluster creation and before test execution.

### Kyma performance test implementation

This section describes Kyma-specific k6 test implementation.

> **NOTE:** For detailed information about the k6 test framework, read the [k6 documentation](https://docs.k6.io).

The Kyma k6 executor has a pre-defined environment variable and tags that provide additional metadata about the current execution and the target test cluster.

> **NOTE:** For more information about K6 tags, read [this](https://docs.k6.io/docs/tags-and-groups) document.

These are the available environment variables:
- **CLUSTER_DOMAIN_NAME** is the domain name of the target Kyma load test cluster.
- **REVISION** is the SHA ID of the tested `master` branch.

These are the pre-defined test execution tags:
- **testName** is the name of a test scenario that every test should provide in the test script implementation.
- **component** is the tested component or area. Provide this tag also with the test script implementation.
- **revision** is the SHA ID of the `master` branch used for tests. It is provided with the Kyma performance test runner. The test script should use the **REVISION** variable as a value.

The tags allow you to distinguish test results in [Grafana](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1).

See [this](./components/examples/http-db-service.js) file for a k6 test example run for **http-db-service**, that contains the pre-defined **testName** and **component** tag names:

The example k6 script below require a test scenario deployment, which will deploy sample service and expose the service API.

Scenario deployment will be triggered from `setup.sh` which places in directory `prerequisites` in the corresponding subdirectory of component, in this case directory `prerequisites\examples`.

The file `setup.sh` will be executed from Kyma load test generator to bootstrap setup process.

```bash
#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

kubectl apply -f ${WORKING_DIR}/example.yaml -n example-test-namespace
```

The example `setup.sh` above trigger a `kubectl apply` command for an example deployment which described in the file `example.yaml` and placed in the same directory as `setup.sh`, here is the important line is the
```bash
WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
```
which will ensure the current working directory for `setup.sh` doesnt matter where you trigger `setup.sh`, if this line not provided and used, you will have some side-effect since `setup.sh` used also from Kyma load generator.

```javascript
import http from 'k6/http';
import { check, sleep } from "k6";

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    tags: {
        "testName": "http_db_service_10vu_60s_1000",
        "component": "http-db-service",
        "revision": `${__ENV.REVISION}`
    }
}

export default function() {
    const response = http.get(`https://http-db-service.${__ENV.CLUSTER_DOMAIN_NAME}/`);

    check(response, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 200
    });
    sleep(1);
}
```

This example executes a **1** minute long load test that runs with **1000** request per second. The test is run against **http-db-service**, on a cluster deployed on **CLUSTER_DOMAIN_NAME**, across **10** virtual users.

The test logic should be implemented in a function defined as **default**.

> **NOTE:** Read more about the test execution lifecycle [here](https://docs.k6.io/docs/test-life-cycle).

The **options** variable defines the execution behavior of the test implementation, where:

- **vus** defines the number of virtual users.
- **duration** defines test execution duration.
- **rps** defines a number of requests per second across all virtual users.
- **tags** defines custom tags to add metadata to test execution.

> **NOTE:** Read [this](https://docs.k6.io/docs/options) document for more details on the available options and the test execution behavior.

The test result is stored in InfluxDB and in Grafana that is deployed on a Kyma load generator cluster. You can access the result [here](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1).

## Usage

To run a k6 test, use this command:
```bash
k6 run github.com/kyma-project/kyma/tests/perf/components/examples/http_basic.js
```

This command runs the script once, fetching the `http_basic.js` file from GitHub and triggering one virtual user.

To run local files, use this command:
```bash
k6 run http_basic.js
```

### Run a load test locally

Although k6 test scripts are designed to run on a Kyma load generator cluster automatically, you can run every load test locally without the Kyma load generator.

See an example that deploys a test component on a Kyma cluster to execute a load test:

>NOTE: To run the test locally, install k6 on your local machine and make sure it is running. See [this](https://docs.k6.io/docs/installation) document for the installation instructions.


1. Deploy an example test service:

```bash
./prerequisites/examples/setup.sh
```

2. After deploying the test service, start the load test locally, with an environment variable that represents the domain name of the Kyma test cluster:

```bash
CLUSTER_DOMAIN_NAME=loadtest.cluster.kyma.cx REVISION=123 k6 run components/examples/http-db-service.js
```

You can also use the **-e** CLI flag for all platforms:

```bash
k6 run -e CLUSTER_DOMAIN_NAME=loadtest.cluster.kyma.cx -e REVISION=123 components/examples/http-db-service.js
```

If you want to use k6 with Grafana locally:

- [Start Influxdb & Grafana](https://k6.io/docs/results-visualization/influxdb-+-grafana)
- Run k6 with `--out influxdb=http://localhost:8086/myk6db` option
