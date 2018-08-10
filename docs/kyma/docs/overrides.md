# DRAFT

Kyma components are packaged as [Helm](https://github.com/helm/helm/tree/master/docs) charts and are installed by the [Installer](../../../components/installer/README.md).
This document describes how to configure Installer with override values for Helm chart templates.

## Overview

The Installer is a Kubernetes Operator that uses Helm golang client to install Kyma components.

Some components must be configured to align with installation environment (e.g. Domain name, TLS Certificates, Api Server CA etc.)
You might also want to customize the installation - for example: Provide different memory setting for particular component or change some database properties.

To address these needs, the Installer provides a generic way to configure overrides for your Helm charts.

There are two sources of overrides for Installer:
- Default overrides: Defined in `overrides.yaml`, a file in the root of installation package, that specifies versions of all kyma-specific components and other static overrides.
- User-provided overrides: Optionally defined by the User. User-provided overrides have precedence over default ones.

**Note**: *Overrides described here are using Helm mechanisms without any changes or customizations, please consult Helm documentation in case of any doubts. What's described here is the way to pass overrides to Kyma Installer in order to use them during Helm install/upgrade operations.*


## User-provided overrides

Installer finds user-provided overrides by reading ConfigMaps and Secrets deployed in `kyma-installer` namespace and marked with `installer:overrides` label.
Installer constructs the override by inspecting the ConfigMap/Secret entry key name. The key name should be a dot-separated sequence of strings, corresponding to the template values used in the chart template. See examples below.

**Note**: *Regardless of how many user-provided override objects (ConfigMaps/Secrets) exist, Installer merges them all into one single override structure and is using that to install all components.*

#### Top-level charts overrides

Overrides for top-level charts are straightforward. Just use template value from the chart (without leading ".Values." prefix) as the entry key in ConfigMap/Secret.

Example:

There's a `core` top-level chart, that is installed by the Installer.
In one of it's templates there's a following line with a value reference:
```
memory: {{ .Values.test.acceptance.ui.requests.memory }}
```
In order to resolve this template, there's a definition for "test.acceptance.ui.requests.memory" in the `values.yaml` file of the chart.
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
When a sub-chart contains `values.yaml` file, all values in this file are resolved against the chart itself, so information about chart location is not necessary there.
But when the Installer installs a chart with it's sub-charts, all template values for a sub-chart must be prefixed with a sub-chart "path" (relative to top-level "parent" chart). This in not Installer-specific requirement, the same considerations apply when you provide overrides manually via Helm CLI.

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

The following fragment of `values.yaml` file in `connector-service` chart defines value for _tokenExpirationMinutes_:
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

