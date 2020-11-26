# k3s-tests

## Overview

This project contains the chart with test Job for function-controller.

## Details

This chart _SHOULD NOT_ be installed with Kyma.

This chart is installed in pre-master-serverless-integration-k3s Prow job. `values.yaml` file has to have the same shape as [values.yaml](../../values.yaml) in parent Chart to properly inject needed fields.
