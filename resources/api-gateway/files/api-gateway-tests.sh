#!/bin/bash

sleep 10

cat << EOF > istio-dr.yaml
---
apiVersion: networking.istio.io/v1alpha3 
kind: DestinationRule
metadata:
  name: istio-ingressgateway
  namespace: istio-system
spec:
  host: istio-ingressgateway.istio-system.svc.cluster.local 
  trafficPolicy:
    tls:
      mode: DISABLE
EOF
{{- if .Values.global.isLocalEnv }}
kubectl apply -f istio-dr.yaml -n istio-system
sleep 5
{{- end }}

./app.test
exit_code=$?

{{- if .Values.global.isLocalEnv }}
kubectl delete -f istio-dr.yaml -n istio-system
{{- end }}

curl -XPOST http://127.0.0.1:15020/quitquitquit
sleep 5

exit $exit_code