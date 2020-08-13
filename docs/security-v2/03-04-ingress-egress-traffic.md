---
title: Ingress ans Egress traffic
type: Details
---

## Ingress
Kyma uses the [Istio Ingress Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/) in order to handle all incoming traffic, manage TLS termination or handle mTLS communication between the cluster and external services. By default a `kyma-gateway` is configured as the main point of entry, which exposes all applications using the supplied domain and certificate.
Exposing applications is done using the [API Gateway](components/api-gateway/#overview-overview) controller. 

The current setup of the main gateway can be found [here](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/gateway/templates/gateway.yaml). Configuration description:
- The gateway listens on port `80` and `443`
- Port `80` is automatically redirected to `443`
- The TLS connection requires at least `TLSV1_1`, meaning `TLSV1_0` is rejected
- Accepted cipher suites: ECDHE-RSA-CHACHA20-POLY1305, ECDHE-RSA-AES256-GCM-SHA384, ECDHE-RSA-AES256-SHA, ECDHE-RSA-AES128-GCM-SHA256, ECDHE-RSA-AES128-SHA

## TLS Managmenet


## Egress
Currently no Egress limitations are implemented, meaning that all applications deployed in the Kyma cluster can access outside resources without limitations.
However, a future improvement is to enable [Istio Egress Gateway](https://istio.io/latest/docs/tasks/traffic-management/egress/egress-gateway/), which can limit the number of external services accessible to the applications running within the cluster.

>**NOTE:** in the case of connection problems with external services it may be required to create an [Service Entry](https://istio.io/latest/docs/reference/config/networking/service-entry/) object to register the service. 

