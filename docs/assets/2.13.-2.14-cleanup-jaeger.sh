#!/usr/bin/env bash

kubectl delete -n kyma-system tracepipelines.telemetry.kyma-project.io jaeger
kubectl delete -n kyma-system telemetries.telemetry.istio.io kyma-traces
kubectl delete -n kyma-system jaegers.jaegertracing.io tracing-jaeger
kubectl delete -n kyma-system serviceaccounts tracing-auth-proxy
kubectl delete -n kyma-system serviceaccounts tracing
kubectl delete -n kyma-system secrets tracing-auth-proxy-default
kubectl delete -n kyma-system secrets jaeger-operator-service-cert
kubectl delete -n kyma-system configmaps tracing-auth-proxy-tracing-templates
kubectl delete -n kyma-system configmaps tracing-grafana-dashboard
kubectl delete -n kyma-system configmaps tracing-grafana-datasource
kubectl delete -n kyma-system clusterroles.rbac.authorization.k8s.io tracing
kubectl delete -n kyma-system clusterrolebindings.rbac.authorization.k8s.io tracing
kubectl delete -n kyma-system services tracing-jaeger-query-secured
kubectl delete -n kyma-system services tracing-jaeger-metrics
kubectl delete -n kyma-system services zipkin
kubectl delete -n kyma-system services tracing-jaeger-operator
kubectl delete -n kyma-system deployments.apps tracing-jaeger-operator
kubectl delete -n kyma-system deployments.apps tracing-auth-proxy
kubectl delete -n kyma-system authorizationpolicies.security.istio.io tracing-jaeger
kubectl delete -n kyma-system peerauthentications.security.istio.io tracing-jaeger-operator-metrics
kubectl delete -n kyma-system peerauthentications.security.istio.io tracing-jaeger-metrics
kubectl delete -n kyma-system virtualservices.networking.istio.io tracing