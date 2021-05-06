# k3s-tests

## Overview

This project contains a chart with the test Job for the Eventing components.

## Details

> **CAUTION:** This chart should not be installed with Kyma.

This chart is installed in the `pre-main-eventing-integration-k3s` Prow job. The `values.yaml` file must have the same shape as [`values.yaml`](../../values.yaml) in the parent chart to properly inject required fields.
