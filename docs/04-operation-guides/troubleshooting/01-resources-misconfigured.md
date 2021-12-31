---
title: Kyma resource is misconfigured
---

All resources in the Kyma-native Namespace with the `reconciler.kyma-project.io/managed-by-reconciler-disclaimer` annotation are managed by Reconciler. However, Reconciler does not support all use cases.

## Symptom

A resource is not working as expected. It could be either invalid, unavailable, etc.

## Cause

A user modified the resource, and it is misconfigured in such a way that Reconciler cannot reconcile it.

## Remedy

Delete the misconfigured resource and wait for Reconciler to reinstall it.
