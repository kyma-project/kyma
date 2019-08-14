# Installer resources

## Overview

This directory contains resources used by the Installer.

## Installer CRs

Files starting with `installer-cr-*` contain different component configurations that you can provide to the Installer.

>**NOTE:** When you add a new component to Kyma, you must update every Installer CR that will use the component.

### Available configurations

These are the currently available Installer configurations:

| Configuration | Description | Part of the release |
|----------------|------|------|
| `installer-cr.yaml.tpl` | Provides complete local Kyma installation. | ✅ |
| `installer-cr-cluster.yaml.tpl` | Provides complete cluster Kyma installation. | ✅ |
| `installer-cr-cluster-with-compass.yaml.tpl` | Provides complete cluster Kyma installation with the `compass` and `compass-runtime-agent` components. These components will eventually become part of complete Kyma installation and this configuration will be deleted.  | ⛔️ |
| `installer-cr-cluster-compass.yaml.tpl` | Provides cluster Compass installation only with the Kyma components that Compass uses. | ⛔️ |
| `installer-cr-cluster-runtime.yaml.tpl` | Provides complete cluster Kyma installation with the `compass-runtime-agent` component. | ⛔️ |
