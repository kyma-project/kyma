---
title: '"FAILED" status for created ServiceInstances'
---

## Symptom

Your ServiceInstance creation was successful and yet the release is marked as `FAILED` on the releases list when running the `helm list` command.

## Cause

There is an error on the Helm's side that was not passed on to the Helm Broker.

## Remedy

To get the error details, check the Helm release status.
