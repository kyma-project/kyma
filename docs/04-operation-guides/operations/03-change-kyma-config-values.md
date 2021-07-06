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

  - For `ory`, the values of `hydra.deployment.resources.limits.cpu` and `hydra.deployment.resources.requests.cpu` will be overriden to `153m` and `53m` respectively.
    
  - For `monitoring`, the values of `alertmanager.alertmanagerSpec.resources.limits.memory` and `alertmanager.alertmanagerSpec.resources.requests.memory` will be overriden to `304Mi` and `204Mi` respectively.
  
  ```yaml
  ory:
    hydra:
      deployment:
        resources:
          limits:
            cpu: 153m
          requests:
            cpu: 53m
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
  kyma deploy --value ory.hydra.deployment.resources.limits.cpu=153m \
  --value ory.hydra.deployment.resources.requests.cpu=53m \
  --value monitoring.alertmanager.alertmanagerSpec.resources.limits.memory=304Mi \
  --value monitoring.alertmanager.alertmanagerSpec.resources.requests.memory=204Mi
  ```

> **NOTE:** If a value is defined several times, the last value definition in the list is used. The `--value` flag also overrides any conflicting value that is defined with a `--value-file` flag.
