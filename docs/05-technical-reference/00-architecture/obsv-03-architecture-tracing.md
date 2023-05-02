---
title: Tracing Architecture
---

With the stratergy shift to more [flexibility](https://blogs.sap.com/2022/09/25/from-observability-to-telemetry-a-strategy-shift-in-sap-btp-kyma-runtime/), a layer or collector is introduced for collecting trace data. This collector pushes the trace data to [configured](../../01-overview/main-areas/telemetry/telemetry-03-traces.md#setting-up-a-tracepipeline) trace backend. The user can bring his own trace backend or [Install custom Jaeger in Kyma](https://github.com/kyma-project/examples/tree/main/jaeger). One can inspect specific traces using the trace backend

## Architecture diagram

![Tracing architecture](./assets/obsv-tracing-architecture.svg)

## Flow: Collect traces

The process of collecting traces by trace backend looks as follows:

1. The application receives a request, either from an internal or external source.
2. If the application has Istio injection enabled and HTTP headers are missing, [Istio proxy](https://github.com/istio/proxy) enriches the request with the correct HTTP headers and propagates them to the Application container. Furthermore, Istio proxy sends the trace data for any intercepted request the otel-collector instance served by the [Telemetry module](./../../01-overview/main-areas/telemetry/README.md), which is configured by default to ship the trace data to trace backend using the OTLP protocol.  
3. Jaeger stores the trace data on a PersistentVolume and makes the trace information available using an API and UI.

## Flow: Query traces

The process of querying traces from Jaeger looks as follows:

1. A Kyma user accesses the Jaeger UI to look for specific traces.
2. The user passes the request to the Jaeger Query service and retrieves the trace information. Learn more about [accessing Jaeger](../../04-operation-guides/security/sec-06-access-expose-grafana.md).
3. The Jaeger Deployment provides trace information.