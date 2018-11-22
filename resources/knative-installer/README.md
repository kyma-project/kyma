# Knative installer

## Overview

This chart packs the [knative installation script](../../components/knative-installer/README.md) as a Kubernetes job.

Installed releases of Knative components may be found in [configmap](./templates/configmap.yaml) as values of `SERVING_URL` and `EVENTING_URL` environmental variables.