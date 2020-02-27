# Serverless Performance Tests

See [Performance Tests](../../README.md) for an introduction to performamce testing in Kyma.

Some notes on the serverless performance test:

- deploys for function size `s`, `m` an api, function and horizontalpodautoscaler
- function sizes `xl` and `l` require too many minimum replicas and therefore too much RAM on the test-cluster, therefore it is commented out for now
- there is a combined k6 script to test all functions: it would be better to have a separate test setup for each function such that autoscaling of functions does not affect each other (e.g. due to limited memory). However there is currently [no way to delete resources after a test](https://github.com/kyma-project/test-infra/issues/1025). We could test one lambda size after another, however this requires to have one k6 file per function which leads to code duplication. Therefore the combined k6 file is chosen
- minimum replicas of deployment is guaranteed by `setup.sh`
- yaml files can be configured via environment variables by leveraging `envsubst`
- to figure out the maximum throughput there are multiple stages: each stage increases the maximum of [virtual users](https://k6.io/docs/getting-started/running-k6)
  - each step has 90s such that autoscaling has enough time to scale up the number of replicas
