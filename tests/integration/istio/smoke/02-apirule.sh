/bin/bash

if [[ -z "${CLUSTER_DOMAIN}" ]]; then
  echo "CLUSTER_DOMAIN is not set"
  exit 3
fi

cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: $NAMESPACE
  spec:
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - accessStrategies:
        - config: {}
          handler: allow
        methods:
          - GET
          - POST
          - PUT
          - PATCH
          - DELETE
          - HEAD
        path: /.*
    service:
      host: httpbin.$CLUSTER_DOMAIN
      name: httpbin
      port: 8000
EOF