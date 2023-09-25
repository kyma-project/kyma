---
title: Change Kyma settings
---

To change your Kyma settings, you simply deploy the same Kyma version that you're currently using, just with different configuration values.

You can use the `--values-file` and the `--value` flag.

- To override the standard Kyma configuration, run:

  ```bash
  kyma deploy --values-file {VALUES_FILE_PATH}
  ```

  In the following example, `{VALUES_FILE_PATH}` is the path to a YAML file containing the desired configuration:

  - For `global`, the values of `images.istio_pilot.version`, `images.istio_pilot.directory` and `containerRegistry.path` will be overridden to `1.11.4`, `istio` and `docker.io` respectively.
  - For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overridden to `304Mi` and `204Mi` respectively.

  ```yaml
  global:
    containerRegistry:
      path: docker.io
    images:
      istio_pilot:
        version: 1.11.4
        directory: "istio"
  monitoring:
    alertmanager:
      alertmanagerSpec:
        resources:
          limits:
            memory: 304Mi
          requests:
            memory: 204Mi
  ```

- You can also provide multiple values files at the same time:

  ```bash
  kyma deploy --values-file {VALUES_FILE_1_PATH} --values-file {VALUES_FILE_2_PATH}
  ```

> **NOTE:** If a value is defined in several files, the value of the last file in the list is used.

- Alternatively, you can specify single values instead of a file:

  ```bash
  kyma deploy
  --value monitoring.alertmanager.alertmanagerSpec.resources.limits.memory=304Mi \
  --value monitoring.alertmanager.alertmanagerSpec.resources.requests.memory=204Mi
  ```

> **NOTE:** If a value is defined several times, the last value definition in the list is used. The `--value` flag also overrides any conflicting value that is defined with a `--value-file` flag.
