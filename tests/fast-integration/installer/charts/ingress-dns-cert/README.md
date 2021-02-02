# Ingress DNS and certificate requests

## Overview

This chart creates the following certificate request:

```
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: kyma-tls-cert
  namespace: istio-system
spec:s
  commonName: '*.{{ .Values.global.domainName }}'
  secretName: kyma-gateway-certs
  secretRef:
    name: kyma-gateway-certs
    namespace: istio-system
```

It also annotates `istio-ingressgateway` with:
```
dns.gardener.cloud/class='garden' 
dns.gardener.cloud/dnsnames='*.$DOMAIN'
```
