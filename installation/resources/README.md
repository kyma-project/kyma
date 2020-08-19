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
| `installer-cr-cluster-runtime.yaml.tpl` | Provides complete cluster Kyma installation with the `compass-runtime-agent` component. | ⛔️ |
