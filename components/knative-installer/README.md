## Knative installer

## Overview

Knative is distributes as single YAML file with resources definitions. `knative-installer` containerizes the process to make it possible to install Knative with Kyma's installer.

## Prerequisites

Knative requires istio in `istio-system` namespace. It must be installed before `isito-kyma-patch`.

## Usage

This section describes how to use the application.

### Configuration

The application accepts two environmental variables: `SERVING_URL` and `EVENTING_URL` which must contain links to Knative serving and eventing releases.