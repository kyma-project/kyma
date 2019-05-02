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

Directory ```test/perf``` contains all performance tests source code.
A Kyma performance test is a K6 test script with or without prerequisites like custom component deployments or configurations.
A Kyma performance test will runs against [Kyma load test cluster](https://github.com/kyma-project/test-infra).

Each subdirectory in the ```tests/perf``` directory defines source code for one test suite and focusing on one component, 
the subdirectory ```prerequisites``` will contains **yaml** files of custom component deployments 
(like custom configuration or custom scenario deployments) if necessary.

Prerequisites directory content will be deployed after load test cluster deployment and before test execution.

### Implementing Kyma performance test



### Run test locally