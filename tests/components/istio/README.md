# Istio component tests

We have chosen to implement component tests with [cucumber/godog](https://github.com/cucumber/godog) library.
Please feel free to take a look at their repository to learn more about the framework.

## Requirements

In order to run these tests, you need to have:
- Kubernetes installed and kubeconfig configured to point on this cluster
- Kyma installed with at least Istio component
- Export environment variables, the only required is `KYMA_PROFILE`

## Environment variables

These environment variables will determine how the tests are run in both prow and your local machine.

Required:
- `KYMA_PROFILE`: Set this environment variable accordingly to the Kyma profile installed on the Kubernetes cluster. These values are: `evaluation` or `production`.
- `EXPORT_RESULT`: Set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON and HTML report. Default value: `false`.

## Executing tests

### Start the test suite

Having prepared environment variables simply run:
`make test`

We have also provided a Make target that will create k3d cluster, install Kyma on it and run the tests:
`make test-k3d`

## Create k3d cluster

To provision k3d cluster on your local machine run:
`make provision-k3d`

## Install Kyma

To install Kyma with Istio component on your cluster run:
`make kyma-istio-deploy`
