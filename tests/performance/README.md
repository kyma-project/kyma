# Kyma Performance Test Guidelines

## Overview
Kyma uses [K6](https://docs.k6.io/docs) for performance and load testing, K6 is a developer oriented well documented load testing tool, easy to 
integrate into Kyma workflow and automation flow.

### Running K6
To execute a very simple k6 test from the command line, you can use this:
```bash
k6 run github.com/kyma-project/kyma/tests/performance/assets/http_basic.js
```

With command above, k6 will fetch the http_het.js file from Github and only fires 1 virtual user which get execute 
the script once.

K6 can also execute local files.
```bash
k6 run script.js
```

## K6 with Kyma