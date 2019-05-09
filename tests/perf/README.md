# Kyma Performance Test Guidelines

## Overview
Kyma uses [K6](https://docs.k6.io) for performance and load testing, K6 is a developer oriented well documented load testing tool, easy to 
integrate into Kyma workflow and automation flow.

### Running K6
To execute a very simple k6 test from the command line, you can use this:
```bash
k6 run github.com/kyma-project/kyma/tests/perf/components/examples/http_basic.js
```

With command above, k6 will fetch the http_basic.js file from Github and only fires 1 virtual user which get execute 
the script once.

K6 can also execute local files.
```bash
k6 run http_basic.js
```

## K6 with Kyma

Directory ```tests/perf``` contains all performance test script source code.
A Kyma performance test is a K6 test script with or without prerequisites e.g. custom test component and/or scenario deployments.
A Kyma performance test will runs against [Kyma load test cluster](https://github.com/kyma-project/test-infra).

Each subdirectory in the ```tests/perf/components``` directory defines source code for one test suite and focusing on one component or area, 
the subdirectory ```prerequisites``` will contains **yaml** files of test component deployments and shell scripts 
(like custom configuration or custom scenario deployments) if necessary. 
Each component should create a subdirectory in ```prerequisites``` to place yaml files and shell scripts, the name of subdirectory should be same as where test script placed.

e.g. Directory structure for components **application-gateway** and **event-bus** should look like following
```
tests
|  
+--- perf
     |   
     |--- components
     |    |
     |    |--- application-gateway // will contain test scripts for application gateway
     |    |
     |    +--- event-bus           // will contain test scripts for event bus
     |      
     +--- prerequisites
          |
          |--- application-gateway // will contain shell scripts or deployement files for application gateway
          |
          +--- event-bus           // will contain shell scripts or deployement files for event bus

``` 
Prerequisites directory content will be deployed after load test cluster creation and before test execution.

### Implementing Kyma performance test

This section will document Kyma specific k6 test implementation, for detailed information about k6 test framework you can 
read from [original documentation](https://docs.k6.io)

Kyma k6 executor has some pre-defined environment variable and tags to provide some additional meta information about 
current execution and target test cluster.

More about K6 tags please read from [here](https://docs.k6.io/docs/tags-and-groups).

Available environment variables
- **CLUSTER_DOMAIN_NAME**, is the domain name of the target Kyma load test cluster
- **REVISION**, the SHA id of master branch being testing

Pre-Defined test execution tags 
- **testName**, is the name of test scenario which every test should provide in test script implementation. 
- **component**, the component or area which currently being tested this tag also should be provided with test script implementation
- **revision**, SHA id of master branch which used for tests, this will be provided with Kyma performance test runner and test script should not provide 

This information will be used later on [grafana](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1) to filter test results

An example k6 test testing example http-db-service defined [here](./prerequisites/examples/example.yaml), with predefined tag name ```testName``` and ```component```

```javascript
import http from 'k6/http';

export let options = {
    vus: 10,
    duration: "1m",
    rps: 1000,
    tags: {
        "testName": "http_basic_10vu_60s_1000",
        "component": "http-db-service",
        "revision": `${__ENV.REVISION}`
    }
}

export default function() {
    const response = http.get(`https://http-db-service.${__ENV.CLUSTER_DOMAIN_NAME}/`);
}
```

Example test above will execute a load test against http-db-service on a cluster deployed on **CLUSTER_DOMAIN_NAME** 
with **10** virtual users, **1** minute long and **1000** request per second across **10** virtual users.

Test logic should be implemented in a function defined as **default**, more about test execution lifecycle please read [here](https://docs.k6.io/docs/test-life-cycle).

Variable ```options``` defines execution behavior of test implementation. With following options

- ```vus``` defines amount of virtual users.
- ```duration``` defines test execution duration.
- ```rps``` defines request per second across all virtual users
- ```tags``` defines custom tags to add meta information to test execution 

More about available options and test execution behavior please read [here](https://docs.k6.io/docs/options).

Result of test execution will be stored in a Influx-DB along with Grafana on Kyma Load Generator Cluster and can be accessed from [here](https://grafana.perf.kyma-project.io/d/ReuNR5Aik/kyma-performance-test-results?orgId=1)

### Run test locally

Each load test can be executed locally without Kyma load generator for development purposes or testing k6 script locally before 
script executed on load generator cluster. 

Example below shown deployment of example test component on a Kyma cluster and execution of simple load test against

>NOTE: Before you start running test locally ensure k6 installed on your local machine and running. Installation instruction available [here](https://docs.k6.io/docs/installation)

Following example will deploy some test component on Kyma cluster to execute load test against.

First deploy example test service which we execute load test against on Kyma cluster

```bash
kubectl apply -f prerequisites/examples/example.yaml
```

After test service deployed we can start load test locally to against Kyma cluster from command line with an environment 
variable which represent domain name of Kyma test cluster

```bash
CLUSTER_DOMAIN_NAME=loadtest.cluster.kyma.cx REVISION=123 k6 run components/examples/http_get.js
```

or we can use ```-e``` CLI flag for all platform

```bash
k6 run -e CLUSTER_DOMAIN_NAME=loadtest.cluster.kyma.cx -e REVISION=123 components/examples/http_get.js
```