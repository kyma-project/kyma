# Istio component tests

Istio component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes installed and kubeconfig configured to point to this cluster
- Istio component installed on cluster ([example component file](https://github.com/kyma-project/test-infra/blob/main/prow/scripts/cluster-integration/kyma-integration-k3d-istio-components.yaml))
- Environment variables exported (the only required environment variable is `KYMA_PROFILE`)

### Environment variables

These environment variables determine how the tests are run on both Prow and your local machine:

- `KYMA_PROFILE` - set this environment variable accordingly to the Kyma profile installed on the Kubernetes cluster. The possible values are `evaluation` or `production`.
- `EXPORT_RESULT` - set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON, and HTML report. The default value is `false`.

## Usage

To start the test suite, run:

```
make test
```

If you don't have a cluster, you can run the tests on your local machine. To do so, run:

```
make test-k3d
```

This command creates a k3d cluster on your local machine, installs Kyma on it, and runs the tests.