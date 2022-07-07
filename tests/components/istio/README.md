# Istio component tests

Istio component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes installed and kubeconfig configured to point to this cluster
- Kyma installed with at least Istio component
- Export environment variables, the only required is `KYMA_PROFILE`

## Environment variables

These environment variables will determine how the tests are run in both prow and your local machine.

- `KYMA_PROFILE`: Set this environment variable accordingly to the Kyma profile installed on the Kubernetes cluster. These values are: `evaluation` or `production`.
- `EXPORT_RESULT`: Set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON and HTML report. Default value: `false`.

## Executing tests

### Start the test suite

Having prepared environment variables simply run:

```make test```

We have also provided a Make target that will create k3d cluster, install Kyma on it and run the tests:

```make test-k3d```

#### Prepare cluster on your local machine

1. Create k3d cluster

```make provision-k3d```

2. Install Kyma with the Istio component on your cluster:

```make kyma-istio-deploy```
