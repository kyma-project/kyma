# Serverless-Bench

The Serverless-Bench image is based on the `locustio/locust` image. It also includes a locust test configuration and a simple script to run the test and output the results in JSON to stdout. The logs are collected via a log sink and are pushed to Google Cloud BigQuery for further analysis.  


## Running Serverless benchmarks Locally

This a quick guide to simplify running the Function Controller benchmarks' tests in a development environment. This can be useful in several scenarios. For example, benchmarking a new runtime implementation.

To run the these tests, apply the following steps:

- Start with a fresh Kyma cluster. The following minimum components are requireed:
```yaml
---
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "monitoring"
  - name: "serverless"
```

- Apply the functions you need to benchmark from the test [function fixtures](./fixtures/functions/). If you need to benchmark a new function runtime or configuration, you should create a similar function manifest with the same naming pattern. 

- Wait for the functions to be ready.
- Apply the [local-test](./fixtures/local-test/) manifests. This will deploy the following resources:
    - A MySQL Pod and service to store the benchmarking data
    - A grafana datasource to read the benchmark data from the mysql backend.
    - A grafana dashboard configuration to present the data.
- Edit the [test job spec](./fixtures/serverless-benchmark-job.yaml) to set `USE_LOCAL_MYSQL` to `true`. You can also use `CUSTOM_FUNCTIONS` to test specific functions, instead of testing the full list.
- Apply the test job. Each test run will create a data point for each defined function.

- Once the first test run is done, you can [access grafana](https://kyma-project.io/docs/kyma/latest/04-operation-guides/security/sec-06-access-expose-kiali-grafana#access-kiali-grafana-and-jaeger) and find the `Serverless Controller Benchmarks` dashboard, and check the data.