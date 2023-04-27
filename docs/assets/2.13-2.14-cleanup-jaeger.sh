#!/usr/bin/env bash

kubectl delete tracepipelines.telemetry.kyma-project.io jaeger --ignore-not-found
kubectl delete -n istio-system telemetries.telemetry.istio.io kyma-traces --ignore-not-found
kubectl delete -n kyma-system jaegers.jaegertracing.io tracing-jaeger --ignore-not-found
kubectl delete -n kyma-system serviceaccounts tracing-auth-proxy --ignore-not-found
kubectl delete -n kyma-system serviceaccounts tracing --ignore-not-found
kubectl delete -n kyma-system secrets tracing-auth-proxy-default --ignore-not-found
kubectl delete -n kyma-system secrets jaeger-operator-service-cert --ignore-not-found
kubectl delete -n kyma-system configmaps tracing-auth-proxy-tracing-templates --ignore-not-found
kubectl delete -n kyma-system configmaps tracing-grafana-dashboard --ignore-not-found
kubectl delete -n kyma-system configmaps tracing-grafana-datasource --ignore-not-found
kubectl delete clusterroles.rbac.authorization.k8s.io tracing --ignore-not-found
kubectl delete clusterrolebindings.rbac.authorization.k8s.io tracing --ignore-not-found
kubectl delete -n kyma-system services tracing-jaeger-query-secured --ignore-not-found
kubectl delete -n kyma-system services tracing-jaeger-metrics --ignore-not-found
kubectl delete -n kyma-system services zipkin --ignore-not-found
kubectl delete -n kyma-system services tracing-jaeger-operator --ignore-not-found
kubectl delete -n kyma-system deployments.apps tracing-jaeger-operator --ignore-not-found
kubectl delete -n kyma-system deployments.apps tracing-auth-proxy --ignore-not-found
kubectl delete -n kyma-system authorizationpolicies.security.istio.io tracing-jaeger --ignore-not-found
kubectl delete -n kyma-system peerauthentications.security.istio.io tracing-jaeger-operator-metrics --ignore-not-found
kubectl delete -n kyma-system peerauthentications.security.istio.io tracing-jaeger-metrics --ignore-not-found
kubectl delete -n kyma-system virtualservices.networking.istio.io tracing --ignore-not-found