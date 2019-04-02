---
title: Testing Kyma
type: Details
---

For testing, the Kyma components use the [Octopus](http://github.com/kyma-incubator/octopus). 
Octopus is a testing framework that allows you to run tests defined as Docker images on a running cluster.
Octopus uses two types of Custom Resource Definitions:
- TestDefinition defines your test as a Pod specification.
- ClusterTestSuite defines a suite of tests to execute and how to execute them.

Place your TestDefinition in your chart.

## Add a new test

See the following example of a test prepared for Dex:

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
This simple test calls the `Dex` endpoint with cURL, defined as follows:

```yaml
apiVersion: "testing.kyma-project.io/v1alpha1"
kind: TestDefinition
metadata:
  name: "test-{{ template "fullname" . }}-connection-dex"
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - name: "test-{{ template "fullname" . }}-connection-dex"
        image: tutum/curl:alpine
        command: ["/usr/bin/curl"]
        args: [
          "--fail",
          "--max-time", "10",
          "--retry", "60",
          "--retry-delay", "3",
          "http://dex-service.{{ .Release.Namespace }}.svc.cluster.local:5556/.well-known/openid-configuration"
        ]
      restartPolicy: Never
---

```

## Test execution
To execute all tests, use the `testing.sh` script located in the `/installation/scripts/` directory. 
Internally, `ClusterTestSuite` resource is defined, that fetches all `TestDefinitions` and executes them.


### Run a test manually
To execute tests manually, create `ClusterTestSuite` on your own. See the following example:

```yaml
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: {my-suite}
spec:
  maxRetries: 0
  concurrency: 1
  count: 1
```

Creation of the suite triggers tests. Current progress of the tests is visible in a `ClusterTestSuite` status.
```bash
 kubectl get cts {my-suite} -oyaml
 ```
 
The sample output looks as follows:
```
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  name: {my-suite}
spec:
  concurrency: 1
  count: 1
  maxRetries: 0
status:
  conditions:
  - status: "True"
    type: Running
  results:
  - executions:
    - completionTime: 2019-04-05T12:23:00Z
      id: {my-suite}-test-dex-dex-connection-dex-0
      podPhase: Succeeded
      startTime: 2019-04-05T12:22:54Z
    name: test-dex-dex-connection-dex
    namespace: kyma-system
    status: Succeeded
  - executions:
    - id: {my-suite}-test-core-core-ui-acceptance-0
      podPhase: Running
      startTime: 2019-04-05T12:37:53Z
    name: test-core-core-ui-acceptance
    namespace: kyma-system
    status: Running
  - executions: []
    name: test-api-controller-acceptance
    namespace: kyma-system
    status: NotYetScheduled
  startTime: 2019-04-05T12:22:53Z
```

The ID of the execution is the same as the ID of the testing Pod. The testing Pod is created in the same namespace as it's TestDefinition. To get logs for a specific test, execute the following command:
```
kubeclt logs {execution-id} -n {test-def-namespace}
```