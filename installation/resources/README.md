# Installer resources

## Overview

This directory contains resources used by the Installer.

## Installer CRs

Files starting with `installer-cr-*` contain different component configurations that you can provide to the Installer. These are the currently available configurations:

| Configuration | Description | Supported |
|----------------|------|------|
| `installer-cr.yaml.tpl` | Provides complete local Kyma installation. | ✅ |
| `installer-cr-cluster.yaml.tpl` | Provides complete cluster Kyma installation. | ✅ |
| `installer-cr-cluster-with-compass.yaml.tpl` | Provides complete cluster Kyma installation with the `compass` and `compass-runtime-agent` components. | ⛔️ |
| `installer-cr-cluster-compass-minimal.yaml.tpl` | Provides cluster Compass installation only with the Kyma components that Compass uses. | ⛔️ |
| `installer-cr-cluster-agent.yaml.tpl` | Provides complete cluster Kyma installation with the `compass-runtime-agent` component. | ⛔️ |

## Add a new component to Kyma

When you add a new component to Kyma, you must update every Installer CR that will use the component.
