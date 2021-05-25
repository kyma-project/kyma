---
title: Testing Kyma
type: Details
---

Kyma components use [Octopus](http://github.com/kyma-incubator/octopus) for testing.
Octopus is a testing framework that allows you to run tests defined as Docker images on a running cluster.
Octopus uses two CustomResourceDefinitions (CRDs):

- TestDefinition, which defines your test as a Pod specification.
- ClusterTestSuite, which defines a suite of tests to execute and how to execute them.

## Add a new test

To add a new test, create a `yaml` file with TestDefinition CR in your chart. To comply with the convention, place it under the `tests` directory.
See the exemplary chart structure for Dex:

```
# Chart tree
dex
├── Chart.yaml
├── README.md
├── templates
│   ├── tests
│   │   └── test-dex-connection.yaml
│   ├── dex-deployment.yaml
│   ├── dex-ingress.yaml
│   ├── dex-rbac-role.yaml
│   ├── dex-service.yaml
│   ├── pre-install-dex-account.yaml
│   ├── pre-install-dex-config-map.yaml
│   └── pre-install-dex-secrets.yaml
└── values.yaml
```

The test adds a new **test-dex-connection.yaml** under the `templates/tests` directory.
For more information on TestDefinition, read the [Octopus documentation](https://github.com/kyma-incubator/octopus/blob/master/docs/crd-test-definition.md).

The following example presents TestDefinition with a container that calls the Dex endpoint with cURL. You must define at least the **spec.template** parameter which is of the `PodTemplateSpec` type.

```yaml
apiVersion: "testing.kyma-project.io/v1alpha1"
kind: TestDefinition
metadata:
  name: {{ .Chart.Name }}
  labels:
    app: {{ .Chart.Name }}-tests
    app.kubernetes.io/name: {{ .Chart.Name }}-tests
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - name: tests
        image: eu.gcr.io/kyma-project/external/curlimages/curl:7.70.0
        command: ["/usr/bin/curl"]
        args: [
          "--fail",
          "--max-time", "10",
          "--retry", "60",
          "--retry-delay", "3",
          "http://dex-service.{{ .Release.Namespace }}.svc.cluster.local:5556/.well-known/openid-configuration"
        ]
      restartPolicy: Never

```

## Tests execution

To run all tests deployed on a Kyma cluster using [Kyma CLI](https://github.com/kyma-project/cli), run:

```bash
kyma test run
```

>**WARNING:** The `kubeconfig` file downloaded from UI does not grant enough privileges to run a test using Kyma CLI. Instead, use the `kubeconfig` file from the cloud provider.

Internally, the ClusterTestSuite resource is defined. It fetches all TestDefinitions and executes them.

### Run tests manually
To run tests manually, you can pass test names to Kyma CLI explicitly. To list all the deployed TestDefinition sets, run:

```bash
kyma test definitions
```

Then, run only the desired tests by passing the TestDefinition names:

```bash
kyma test run <test-definition-1> <test-definition-2> ...
```

See the current tests progress in the ClusterTestSuite status. Run:

```bash
kyma test status
```

The ID of the test execution is the same as the ID of the testing Pod. The testing Pod is created in the same Namespace as its TestDefinition. To get logs for a specific test, run the following command:

```bash
kyma test logs <test-suite-1> <test-suite-2> ...
```
