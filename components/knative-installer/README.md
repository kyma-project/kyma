## Knative installer

## Overview

Knative is distributed as a single YAML file containing resource definitions. The Knative installer containerizes the process to allow the installation of Knative using the Kyma installer.

## Prerequisites

Knative requires running Istio in the `istio-system` Namespace. It must be installed before `isito-kyma-patch` is applied.
