# Serverless-Bench

The Serverless-Bench image is based on the `locustio/locust` image. It also includes a locust test configuration and a simple script to run the test, and output the results in JSON to stdout. The logs are collected with a log sink and are pushed to Google Cloud BigQuery for further analysis.  

## Running Serverless benchmarks Locally

This a quick guide to simplify running the Function Controller benchmarks' tests in a development environment. This can be useful in several scenarios, such as benchmarking a new runtime implementation.

To run the these tests, apply the following steps:

1. Start with a fresh Kyma cluster. The following minimum components are required:

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

2. Apply the Functions you need to benchmark from the test the [Function fixtures](./fixtures/functions/). If you need to benchmark a new Function runtime or configuration, you must create a similar Function manifest with the same naming pattern.

3. Wait for the Functions to be ready.
4. Apply the [local-test](./fixtures/local-test/) manifests. This will deploy the following resources:
    - A MySQL Pod and service to store the benchmarking data
    - A Grafana datasource to read the benchmark data from the MySQL backend.
    - The Grafana dashboard configuration to present the data.
5. Edit the [test job spec](./fixtures/serverless-benchmark-job.yaml) to set `USE_LOCAL_MYSQL` to `true`. You can also use `CUSTOM_FUNCTIONS` to test specific Functions instead of testing the full list.
6. Apply the test job. Each test run creates a data point for each defined Function.

7. Once the first test run is done, you can [access Grafana](https://kyma-project.io/#/04-operation-guides/security/sec-06-access-expose-grafana) and find the `Serverless Controller Benchmarks` dashboard to check the data.
