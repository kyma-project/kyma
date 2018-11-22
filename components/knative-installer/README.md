## Knative installer

## Overview

Knative is distributes as single YAML file with resources definitions. `knative-installer` containerizes the process to make it possible to install Knative with Kyma's installer.

## Prerequisites

Knative requires istio in `istio-system` namespace. It must be installed before `isito-kyma-patch`.
