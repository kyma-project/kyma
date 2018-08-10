---
title: Helm overrides for Kyma installation
type: Getting Started
---

Kyma components are packaged as [Helm](https://github.com/helm/helm/tree/master/docs) charts and are installed by the [Installer](../../../components/installer/README.md).
This document describes how to configure Installer with overrides for Helm [charts](https://github.com/helm/helm/blob/master/docs/charts.md).


## Overview

The Installer is a Kubernetes Operator that uses Helm to install Kyma components.
In order to customize installation for particular environment, one can use Helm *overrides* mechanism.
When using Installer for Kyma installation, users can't interact with Helm directly (installation is not an interactive process).
Instead, the Installer exposes a generic mechanism for providing user-defined overrides.

## User-defined overrides

Installer finds user-defined overrides by reading ConfigMaps and Secrets deployed in `kyma-installer` namespace and marked with `installer:overrides` label.
Installer constructs the override by inspecting the ConfigMap/Secret entry key name. The key name should be a dot-separated sequence of strings, corresponding to the structure of keys in chart values.yaml file. See Examples below.

### Common vs component overides

Installer looks for available overrides each time a component installation/update operation is due.
Overrides for the component are composed from two sets: common overrides and component-specific overrides.
- Common overrides are used for installation of all components. They are defined in ConfigMaps/Secrets marked with `installer:overrides` label only.
- Component-specific overrides are used only for the installation of the component. They are defined in ConfigMaps/Secrets marked with both: `installer:overrides` and `component: <name>` labels, where `<name>` is the component name. Component-Specific overrides have precedence over Common ones.

### Examples

#### Top-level charts overrides

Overrides for top-level charts are straightforward. Just use template value from the chart (without leading ".Values." prefix) as the entry key in ConfigMap/Secret.

Example:

There's a `core` top-level chart, that is installed by the Installer.
In one of it's templates there's a following line with a value reference:
```
memory: {{ .Values.test.acceptance.ui.requests.memory }}
```
In order to resolve this template, there's a default value for "test.acceptance.ui.requests.memory" in the `values.yaml` file of the chart.
The following fragment of `values.yaml` shows this definition:
```
test:
  acceptance:
    ui:
      requests:
        memory: "1Gi"
```

If you want to override this value, for example to "2Gi", proceed as follows:
- Create a ConfigMap in `kyma-installer` namespace, labelled with: `installer:overrides` (or reuse an existing one)
- Add an entry `test.acceptance.ui.requests.memory: 2Gi` to the map.

Once the installation starts, Installer generates overrides based on the map entries and value of "2Gi" will be used instead of default "1Gi" from the chart `values.yaml` file.
For values that should be kept in Secrets, just define a Secret object instead of ConfigMap with the same key and value. Don't forget to label the Secret with `installer:overrides` and remember that values in Secrets must be base64-encoded.


#### Sub-chart overrides

Overrides for sub charts follow the same convention as top-level charts, but require additional information about sub-chart location.
When a sub-chart contains `values.yaml` file, the information about chart location is not necessary because chart and it's `values.yaml` file are on the same level in directory hierarchy.
The situation is different when the Installer installs a chart with sub-charts.
All template values for a sub-chart must be prefixed with a sub-chart "path" (relative to top-level "parent" chart).
This in not Installer-specific requirement, the same considerations apply when you provide overrides manually via `helm` command-line tool.

Example:

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

There are important parameters that are shared across the charts, like Domain name.
These are defined using `global` prefix.

For example, to define `global.domain` just use that as key name in ConfigMap/Secret for Installer.
Once the installation starts, Installer is parsing all map entries and collect all global entries together for the purpose of installation.


## Values and types

Installer generally recognizes all override values as **strings** and internally render overrides to Helm as YAML stream with only **string** values.
There is one exception to this rule with respect to handling booleans: Whenever a string "true" or "false" is encountered, it's converted to a corresponding boolean value (true/false).

