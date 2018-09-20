```
       _                            
      | |                           
      | | __ _  ___  __ _  ___ _ __
  _   | |/ _` |/ _ \/ _` |/ _ \ '__|
 | |__| | (_| |  __/ (_| |  __/ |   
  \____/ \__,_|\___|\__, |\___|_|   
                     __/ |          
                    |___/           
```

## Overview
[Jaeger](http://jaeger.readthedocs.io/en/latest/) is a monitoring and tracing tool for microservices-based distributed systems.

## Details
Jaeger installs as an Istio sub-chart. For dependency declaration, see the [requirements.yaml](../../requirements.yaml) file. The Envoy sidecar uses Jaeger to trace the request flow in the Istio service mesh. The communication from Istio and Envoy uses the Zipkin protocol. Jaeger provides compatibility with the Zipkin protocol. This allows you to use Zipkin protocol and clients in Istio, Envoy, and the Kyma services.

For more details, see the [Istio Distributed Tracing](https://istio.io/docs/tasks/telemetry/distributed-tracing.html) documentation.
