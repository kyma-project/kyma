# Installer resources

## Overview

This directory contains resources used by installer.

## Installer CRs

Files starting with `installer-cr-*` contain different component configurations that can be provided to installer. Currently available configurations:

| Configuration | Description | Officially supported |
|----------------|------|------|
| `installer-cr.yaml.tpl` | Complete local Kyma installation | ✅ |
| `installer-cr-cluster.yaml.tpl` | Complete cluster Kyma installation | ✅ |
| `installer-cr-cluster-with-compass.yaml.tpl` | Complete cluster Kyma installation with `compass` and `compass-runtime-agent` components | ⛔️ |
| `installer-cr-cluster-compass-minimal.yaml.tpl` | Cluster Compass installation only with Kyma components used by Compass | ⛔️ |
| `installer-cr-cluster-agent.yaml.tpl` | Complete cluster Kyma installation with `compass-runtime-agent` component | ⛔️ |

### Adding new components to Kyma

When adding new components to Kyma remember to update every Installer CR that should use the new component.
