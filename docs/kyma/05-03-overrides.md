---
title: Helm overrides for Kyma installation
type: Configuration
---

Kyma packages its components into [Helm](https://helm.sh/docs/) charts that the [Kyma Operator](https://github.com/kyma-project/kyma/tree/master/components/kyma-operator) uses during installation and updates.
This document describes how to configure the Kyma Installer with new values for Helm [charts](https://v2.helm.sh/docs/developing_charts/) to override the default settings in `values.yaml` files.

## Overview

The Kyma Operator is a [Kubernetes Operator](https://coreos.com/operators/) that uses Helm to install Kyma components.
Helm provides an **overrides** feature to customize the installation of charts, for example to configure environment-specific values.
When using Kyma Operator for Kyma installation, users can't interact with Helm directly. The installation is not an interactive process.

To customize the Kyma installation, the Kyma Operator exposes a generic mechanism to configure Helm overrides called **user-defined** overrides.

## User-defined overrides

The Kyma Operator finds user-defined overrides by reading the ConfigMaps and Secrets deployed in the `kyma-installer` Namespace and marked with:
- the `installer: overrides` label
- a `component: {COMPONENT_NAME}` label if the override refers to a specific component

>**NOTE:** There is also an additional `kyma-project.io/installation: ""` label in all ConfigMaps and Secrets that allows you to easily filter the installation resources.

The Kyma Operator constructs a single override by inspecting the ConfigMap or Secret entry key name. The key name should be a dot-separated sequence of strings corresponding to the structure of keys in the chart's `values.yaml` file or the entry in chart's template.

The Kyma Operator merges all overrides recursively into a single `yaml` stream and passes it to Helm during the Kyma installation and upgrade operations.

## Common vs. component overrides

The Kyma Operator looks for available overrides each time a component installation or an update operation is due.
Overrides for a component are composed of two sets: **common** overrides and **component-specific** overrides.

Kyma uses common overrides for the installation of all components. ConfigMaps and Secrets marked with the `installer: overrides` label contain the definition.

Kyma uses component-specific overrides only for the installation of specific components. ConfigMaps and Secrets marked with both `installer: overrides` and `component: {component-name}` labels contain the definition. Component-specific overrides have precedence over common ones in case of conflicting entries.

>**NOTE:** Add the additional `kyma-project.io/installation: ""` label to both common and component-specific overrides to enable easy installation resources filtering.

## Overrides examples

### Top-level charts overrides

Overrides for top-level charts are straightforward. Just use the template value from the chart as the entry key in the ConfigMap or Secret. Leave out the `.Values.` prefix.

See the example:

The Installer uses an `asset-store` top-level chart that contains a template with the following value reference:

```
resources: {{ toYaml .Values.resources | indent 12 }}
```

The chart's default values `minio.resources.limits.memory` and `minio.resources.limits.cpu` in the `values.yaml` file resolve the template.
The following fragment of `values.yaml` shows this definition:
```
minio:
  resources:
    limits:
      memory: "128Mi"
      cpu: "100m"
```

To override these values, for example to `512Mi` and `250m`, proceed as follows:
- Create a ConfigMap in the `kyma-installer` Namespace and label it.
- Add the `minio.resources.limits.memory: 512Mi` and `minio.resources.limits.cpu: 250m` entries to the ConfigMap and apply it:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: assetstore-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: assetstore
    kyma-project.io/installation: ""
data:
  minio.resources.limits.memory: 512Mi #increased from 128Mi
  minio.resources.limits.cpu: 250m #increased from 100m
EOF
```

Once the installation starts, the Kyma Operator generates overrides based on the ConfigMap entries. The system uses the values of `512Mi` instead of the default `128Mi` for MinIO memory and `250m` instead of `100m` for MinIO CPU from the chart's `values.yaml` file.

For overrides that the system should keep in Secrets, just define a Secret object instead of a ConfigMap with the same key and a base64-encoded value. Be sure to label the Secret.

If you add the overrides in the runtime, trigger the update process using this command:

```
kubectl label installation/kyma-installation action=install
```

### Sub-chart overrides

Overrides for sub-charts follow the same convention as top-level charts. However, overrides require additional information about sub-chart location.

When a sub-chart contains the `values.yaml` file, the information about the chart location is not necessary because the chart and its `values.yaml` file are on the same level in the directory hierarchy.

The situation is different when the Kyma Operator installs a chart with sub-charts.
All template values for a sub-chart must be prefixed with a sub-chart "path" that is relative to the top-level "parent" chart.

This is not a Kyma Operator-specific requirement. The same considerations apply when you provide overrides manually using the `helm` command-line tool.

For example, there's the `connector-service` sub-chart nested in the `application-connector` chart installed by default as part of the [Kyma package](`connector-service` sub-chart).
In its `deployment.yaml`, there's the following fragment:

```
spec:
  serviceAccountName: {{ .Chart.Name }}
  containers:
  - name: {{ .Chart.Name }}
    image: {{ .Values.global.containerRegistry.path }}/{{ .Values.global.connector_service.dir }}connector-service:{{ .Values.global.connector_service.version }}
    imagePullPolicy: {{ .Values.deployment.image.pullPolicy }}
    args:
      ...
      - "--appTokenExpirationMinutes={{ .Values.deployment.args.appTokenExpirationMinutes }}"
```

This fragment of the `values.yaml` file in the `connector-service` chart defines the default value for `appTokenExpirationMinutes`:

```
deployment:
  ...
  args:
    ...
    appTokenExpirationMinutes: 5
```

To override this value and change it from `5` to `10`, do the following:

- Create a ConfigMap in the `kyma-installer` Namespace and label it.
- Name it after the main component chart in the `resources` folder and add the `-overrides` suffix to it. In this example, that would be `application-connector-overrides`.
- Add the `connector-service.deployment.args.appTokenExpirationMinutes: 10` entry to the ConfigMap.

Notice that the user-provided override key now contains two parts:

- The chart "path" inside the top-level `application-connector` chart called `connector-service`
- The original template value reference from the chart without the `.Values.` prefix, `deployment.args.appTokenExpirationMinutes`.

Once the installation starts, the Kyma Operator generates overrides based on the ConfigMap entries. The system uses the value of `10` instead of the default value of `5` from the `values.yaml` chart file.

## Global overrides

There are several important parameters usually shared across the charts.
Helm convention to provide these requires the use of the `global` override key.
For example, to define the `global.domain` override, just use `global.domain` as the name of the key in a ConfigMap or Secret for the Kyma Operator.

Once the installation starts, the Kyma Operator merges all of the ConfigMap entries and collects all of the global entries under the `global` top-level key to use for the installation.


## Values and types

The Kyma Operator generally recognizes all override values as strings. It internally renders overrides to Helm as a `yaml` stream with only string values.

There is one exception to this rule with respect to handling booleans:
The system converts `true` or `false` strings that it encounters to a corresponding boolean `true` or `false` value.


## Merging and conflicting entries

When the Kyma Operator encounters two overrides with the same key prefix, it tries to merge them.
If both of them represent a ConfigMap (they have nested sub-keys), their nested keys are recursively merged.
If at least one of keys points to a final value, the Kyma Operator performs the merge in a non-deterministic order, so either one of the overrides is rendered in the final `yaml` data.

It is important to avoid overrides having the same keys for final values.


### Non-conflicting merge example

Two overrides with a common key prefix ("a.b"):

```
"a.b.c": "first"
"a.b.d": "second"
```

The Kyma Operator yields the correct output:

```
a:
  b:
    c: first
    d: second
```

### Conflicting merge example

Two overrides with the same key ("a.b"):

```
"a.b": "first"
"a.b": "second"
```

The Kyma Operator yields either:

```
a:
  b: "first"
```

Or (due to non-deterministic merge order):

```
a:
  b: "second"
```
