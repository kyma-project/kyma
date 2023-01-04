# k3s-tests

## Overview

This project contains a chart with the test Job for the Function Controller.

## Details

> **CAUTION:** Do not install this chart with Kyma.

This chart is installed in the `pre-main-serverless-integration-k3s` ProwJob. The `values.yaml` file must have the same shape as [`values.yaml`](../../values.yaml) in the parent chart to properly inject required fields.
