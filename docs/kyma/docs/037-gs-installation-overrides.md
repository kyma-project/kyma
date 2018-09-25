---
title: Helm overrides for Kyma installation
type: Getting Started
---

Kyma packages its components into [Helm](https://github.com/helm/helm/tree/master/docs) charts that the [Installer](../../../components/installer/README.md) uses.
This document describes how to configure the Installer with override values for Helm [charts](https://github.com/helm/helm/blob/master/docs/charts.md).


## Overview

The Installer is a Kubernetes Operator that uses Helm to install Kyma components.
Helm provides an *overrides* feature to customize the installation of charts, such as to configure environment-specific values.
When using Installer for Kyma installation, users can't interact with Helm directly (installation is not an interactive process).
To customize Kyma installation, Installer exposes a generic mechanism for configuring Helm overrides called **user-defined** overrides.


## User-defined overrides

The Installer finds user-defined overrides by reading the ConfigMaps and Secrets deployed in the `kyma-installer` Namespace and marked with the `installer:overrides` Label.

The Installer constructs a single override by inspecting the ConfigMap or Secret entry key name. The key name should be a dot-separated sequence of strings corresponding to the structure of keys in the chart's `values.yaml` file or the entry in chart's template. See the examples below.

Installer merges all overrides recursively into a single YAML stream and passes it to Helm during the Kyma installation/upgrade operation.


### Common vs component overides

The Installer looks for available overrides each time a component installation or update operation is due.
Overrides for the component are composed from two sets: **common** overrides and **component-specific** overrides.

Kyma uses common overrides for the installation of all components. Only ConfigMaps or Secrets marked with the installer:overrides Label contain the definition of overrides (TODO: Fix it).

Kyma uses component-specific overrides only for the installation of specific components. ConfigMaps and Secrets marked with both installer:overrides and component: <name> Labels, where <name> is the component name, contain the definition. Component-specific overrides have precedence over Common ones in case of conflicting entries.


### Examples


#### Top-level charts overrides

Overrides for top-level charts are straightforward. Just use the template value from the chart (without leading ".Values." prefix) as the entry key in the ConfigMap or Secret.

Example:

The Installer uses a `core` top-level chart that contains a template with the following value reference:
```
memory: {{ .Values.test.acceptance.ui.requests.memory }}
```
The chart's default value `test.acceptance.ui.requests.memory` in the `values.yaml` file resolves the template.
The following fragment of `values.yaml` shows this definition:
```
test:
  acceptance:
    ui:
      requests:
        memory: "1Gi"
```

To override this value, for example to "2Gi", proceed as follows:
- Create a ConfigMap in the `kyma-installer` Namespace, labelled with: `installer:overrides` (or reuse an existing one).
- Add an entry `test.acceptance.ui.requests.memory: 2Gi` to the map.

Once the installation starts, the Installer generates overrides based on the map entries. The system uses the value of "2Gi" instead of the default "1Gi" from the chart `values.yaml` file.

For overrides that the system should keep in Secrets, just define a Secret object instead of a ConfigMap with the same key and value. Don't forget to label the Secret with `installer:overrides`. Remember that values in Secrets must be base64-encoded. (TODO: same key and value, but base64...)


#### Sub-chart overrides



Overrides for sub-charts follow the same convention as top-level charts. However, overrides require additional information about sub-chart location.

When a sub-chart contains the `values.yaml` file, the information about the chart location is not necessary because the chart and it's `values.yaml` file are on the same level in the directory hierarchy.

The situation is different when the Installer installs a chart with sub-charts.
All template values for a sub-chart must be prefixed with a sub-chart "path" that is relative to the top-level "parent" chart.

This is not an Installer-specific requirement. The same considerations apply when you provide overrides manually using the `helm` command-line tool.

Here is an example.

There's a `core` top-level chart, that is installed by the Installer.
There's an `application-connector` sub-chart in `core` with another nested sub-chart: `connector-service`.
In one of it's templates there's a following fragment (shortened):
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

The following fragment of `values.yaml` file in `connector-service` chart defines default value for `tokenExpirationMinutes`:
```
deployment:
  args:
    tokenExpirationMinutes: 60
```

If you want to override this value, for example to "90", do as follows:
- Create a ConfigMap in kyma-installer namespace, labelled with: `installer:overrides` (or reuse existing one)
- Add an entry `application-connector.connector-service.deployment.args.tokenExpirationMinutes: 90` to the map.

Notice that the user-provided override key is now composed from two parts:
  - Chart "path" inside top-level `core` chart: **application-connector.connector-service**
  - Original template value reference from the chart (without .Values. prefix): **deployment.args.tokenExpirationMinutes**
Once the installation starts, Installer generates overrides based on the map entries and value of "90" will be used instead of default "60" from the chart `values.yaml` file.


#### Global overrides

Some important parameters are usually shared across the charts.
Helm convention to provide these is by using `global` override key.

For example, to define `global.domain` override, just use "global.domin" as the name of the key in ConfigMap/Secret for Installer.
Once the installation starts, Installer is merging all map entries and collect all global entries together (under `global` top-level key) for the purpose of installation.


## Merging and conflicting entries

When installer encounters two overrides with the same key prefix, it tries to merge them.
If both of them represent a map (they have nested sub-keys), their nested keys are recursively merged.
If at least one of keys points to a final value,  Installer performs the merge in non-deterministic order, so either one or the other override is rendered to the final YAML.
Considering that, **avoid overrides having the same keys for final values**.

### Example of non-conflicting merge:

Having two overrides with a common *key prefix* ("a.b"):

```
"a.b.c": "first"
"a.b.d": "second"
```

Installer yields correct output:

```
a:
  b:
    c: first
    d: second
```

### Example of conflicting merge:

Having two overrides with *the same* key ("a.b"):

```
"a.b": "first"
"a.b": "second"
```

Installer yields either:

```
a:
  b: "first"
```

Or (due to non-deterministic merge order):

```
a:
  b: "second"
```


## Values and types

Installer generally recognizes all override values as **strings** and internally render overrides to Helm as YAML stream with only **string** values.
There is one exception to this rule with respect to handling booleans: Whenever a string "true" or "false" is encountered, it's converted to a corresponding boolean value (true/false).

