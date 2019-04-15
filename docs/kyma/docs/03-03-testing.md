---
title: Testing Kyma
type: Details
---

For testing, the Kyma components use the Helm test concept. Place your test under the `templates` directory as a Pod definition that specifies a container with a given command to run.

## Add a new test

The system bases tests on the Helm broker concept with one modification: adding a Pod label. Before you create a test, see the official [Chart Tests](https://github.com/kubernetes/helm/blob/release-2.10/docs/chart_tests.md) documentation. Then, add the `"helm-chart-test": "true"` label to your Pod template.

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
apiVersion: v1
kind: Pod
metadata:
  name: "test-{{ template "fullname" . }}-connection-dex"
  annotations:
    "helm.sh/hook": test-success
  labels:
      "helm-chart-test": "true" # ! Our customization
spec:
  hostNetwork: true
  containers:
  - name: "test-{{ template "fullname" . }}-connection-dex"
    image: tutum/curl:alpine
    command: ["/usr/bin/curl"]
    args: [
      "--fail",
      "http://dex-service.{{ .Release.Namespace }}.svc.cluster.local:5556/.well-known/openid-configuration"
    ]
  restartPolicy: Never
```

## Test execution

All tests created for charts under `/resources/core/` run automatically after starting Kyma.
If any of the tests fails, the system prints the Pod logs in the terminal, then deletes all the Pods.

>**NOTE:** If you run Kyma locally, by default, the system does not take into account the test's exit code. As a result, the system does not terminate Kyma Docker container, and you can still access it.
To force a termination in case of failing tests, use `--exit-on-test-fail` flag when executing `run.sh` script.

CI propagates the exit status of tests. If any test fails, the whole CI job fails as well.

Follow the same guidelines to add a test which is not a part of any `core` component. However, for test execution, see the **Run a test manually** section in this document.

### Run a test manually

To run a test manually, use the `testing.sh` script located in the `/installation/scripts/` directory which runs all tests defined for `core` releases.
If any of the tests fails, the system prints the Pod logs in the terminal, then deletes all the Pods.

Another option is to run a Helm test directly on your release.

```bash
$ helm test {your_release_name}
```

You can also run your test on custom releases. If you do this, remember to always delete the Pods after a test ends.
