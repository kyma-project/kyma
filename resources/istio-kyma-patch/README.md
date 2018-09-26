# Istio Kyma patch

## Overview

This chart packs [patch script](../../components/istio-kyma-patch/README.md) as kubernetes job.

## Patches

By default following patches are applied:
 * CRD `policies.authentication.istio.io` is required. This means that security in istio must be enabled.
 * Configuration of sidecar injector ([see more](../../components/istio-kyma-patch/README.md))
 * egressgateway, ingressgateway, policy and telemetry are configured to use zipkin from kyma-system namespace
 * pilot have [kyma webhook](./scripts/webhook.lua) added  
 * monitoring and tracing related resources are deleted