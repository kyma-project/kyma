---
title: Helm overrides for Kyma installation
type: Configuration
---

Kyma packages its components into [Helm](https://github.com/helm/helm/tree/master/docs) charts that the [Installer](https://github.com/kyma-project/kyma/tree/master/components/installer) uses during installation and updates.
This document describes how to configure the Installer with new values for Helm [charts](https://github.com/helm/helm/blob/master/docs/charts.md) to override the default settings in `values.yaml` files.

## Overview

The Installer is a [Kubernetes Operator](https://coreos.com/operators/) that uses Helm to install Kyma components.
Helm provides an **overrides** feature to customize the installation of charts, for example to configure environment-specific values.
When using Installer for Kyma installation, users can't interact with Helm directly. The installation is not an interactive process.

To customize the Kyma installation, the Installer exposes a generic mechanism to configure Helm overrides called **user-defined** overrides.

## User-defined overrides

The Installer finds user-defined overrides by reading the ConfigMaps and Secrets deployed in the `kyma-installer` Namespace and marked with:
- the `installer: overrides` label
- a `component: {COMPONENT_NAME}` label if the override refers to a specific component

>**NOTE:** There is also an additional `kyma-project.io/installation: ""` label in all ConfigMaps and Secrets that allows you to easily filter the installation resources.

The Installer constructs a single override by inspecting the ConfigMap or Secret entry key name. The key name should be a dot-separated sequence of strings corresponding to the structure of keys in the chart's `values.yaml` file or the entry in chart's template.

The Installer merges all overrides recursively into a single `yaml` stream and passes it to Helm during the Kyma installation and upgrade operations.

## Common vs. component overrides

The Installer looks for available overrides each time a component installation or an update operation is due.
Overrides for a component are composed of two sets: **common** overrides and **component-specific** overrides.

Kyma uses common overrides for the installation of all components. ConfigMaps and Secrets marked with the `installer: overrides` label contain the definition.

Kyma uses component-specific overrides only for the installation of specific components. ConfigMaps and Secrets marked with both `installer: overrides` and `component: {component-name}` labels contain the definition. Component-specific overrides have precedence over common ones in case of conflicting entries.

>**NOTE:** Add the additional `kyma-project.io/installation: ""` label to both common and component-specific overrides to enable easy installation resources filtering.

## Overrides examples

### Top-level charts overrides

Overrides for top-level charts are straightforward. Just use the template value from the chart as the entry key in the ConfigMap or Secret. Leave out the `.Values.` prefix.

Se an example:

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

Once the installation starts, the Installer generates overrides based on the ConfigMap entries. The system uses the values of `512Mi` instead of the default `128Mi` for Minio memory and `250m` instead of `100m` for Minio CPU from the chart's `values.yaml` file.

For overrides that the system should keep in Secrets, just define a Secret object instead of a ConfigMap with the same key and a base64-encoded value. Be sure to label the Secret.

If you add the overrides in the runtime, trigger the update process using this command:

```
kubectl label installation/kyma-installation action=install
```

### Sub-chart overrides

Overrides for sub-charts follow the same convention as top-level charts. However, overrides require additional information about sub-chart location.

When a sub-chart contains the `values.yaml` file, the information about the chart location is not necessary because the chart and it's `values.yaml` file are on the same level in the directory hierarchy.

The situation is different when the Installer installs a chart with sub-charts.
All template values for a sub-chart must be prefixed with a sub-chart "path" that is relative to the top-level "parent" chart.

This is not an Installer-specific requirement. The same considerations apply when you provide overrides manually using the `helm` command-line tool.

Here is an example.
There's a `core` top-level chart that the Installer installs.
There's an `application-connector` sub-chart in `core` with a nested `connector-service` sub-chart.
In one of its templates, there's a following fragment:

```
spec:
  containers:
  - name: {{ .Chart.Name }}
	args:
	  - "/connectorservice"
	  - '--appName={{ .Chart.Name }}'
	  - "--domainName={{ .Values.global.domainName }}"
	  - "--tokenExpirationMinutes={{ .Values.deployment.args.tokenExpirationMinutes }}"
```

This fragment of the `values.yaml` file in the `connector-service` chart defines the default value for `tokenExpirationMinutes`:

```
deployment:
  args:
    tokenExpirationMinutes: 60
```

To override this value, and change it from `60` to `90`, do the following:

- Create a ConfigMap in the `kyma-installer` Namespace and label it.
- Add the `application-connector.connector-service.deployment.args.tokenExpirationMinutes: 90` entry to the ConfigMap.

Notice that the user-provided override key now contains two parts:

- The chart "path" inside the top-level `core` chart called `application-connector.connector-service`
- The original template value reference from the chart without the `.Values.` prefix, `deployment.args.tokenExpirationMinutes`.

Once the installation starts, the Installer generates overrides based on the ConfigMap entries. The system uses the value of `90` instead of the default value of `60` from the `values.yaml` chart file.


## Global overrides

There are several important parameters usually shared across the charts.
Helm convention to provide these requires the use of the `global` override key.
For example, to define the `global.domain` override, just use `global.domain` as the name of the key in a ConfigMap or Secret for the Installer.

Once the installation starts, the Installer merges all of the ConfigMap entries and collects all of the global entries under the `global` top-level key to use for the installation.


## Values and types

The Installer generally recognizes all override values as strings. It internally renders overrides to Helm as a `yaml` stream with only string values.

There is one exception to this rule with respect to handling booleans:
The system converts `true` or `false` strings that it encounters to a corresponding boolean `true` or `false` value.


## Merging and conflicting entries

When the Installer encounters two overrides with the same key prefix, it tries to merge them.
If both of them represent a ConfigMap (they have nested sub-keys), their nested keys are recursively merged.
If at least one of keys points to a final value, the Installer performs the merge in a non-deterministic order, so either one of the overrides is rendered in the final `yaml` data.

It is important to avoid overrides having the same keys for final values.


### Non-conflicting merge example

Two overrides with a common key prefix ("a.b"):

```
"a.b.c": "first"
"a.b.d": "second"
```

The Installer yields the correct output:

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

The Installer yields either:

```
a:
  b: "first"
```

Or (due to non-deterministic merge order):

```
a:
  b: "second"
```
