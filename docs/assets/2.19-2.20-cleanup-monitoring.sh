#!/usr/bin/env bash

export DELETE_TIMEOUT="300s"

# Delete Prometheus CRs
kubectl delete -n kyma-system alertmanagers.monitoring.coreos.com monitoring-alertmanager --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheuses.monitoring.coreos.com monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-alertmanager --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-apiserver --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-kubelet --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-prometheus-istio-server-server --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-prometheus-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-alertmanager.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-config-reloaders --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-general.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-k8s.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-apiserver-availability.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-apiserver.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-prometheus-general.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-prometheus-node-recording.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubelet.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-apps --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-resources --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-storage --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-system --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-system-apiserver --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kubernetes-system-kubelet --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-node-exporter.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-node-network --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-node.rules --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-prometheus-operator --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete Deployments
kubectl delete -n kyma-system deployments.apps monitoring-auth-proxy-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system deployments.apps monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system deployments.apps monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system deployments.apps monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system deployments.apps monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT

#Delete Daemonsets
kubectl delete -n kyma-system daemonsets.apps monitoring-prometheus-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system daemonsets.apps monitoring-prometheus-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT


# Delete Services
kubectl delete -n kyma-system services monitoring-alertmanager --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-grafana-secured --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system services monitoring-prometheus-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete Clusterrolebindings
kubectl delete clusterrolebindings.rbac.authorization.k8s.io monitoring-grafana-clusterrolebinding --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterrolebindings.rbac.authorization.k8s.io monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterrolebindings.rbac.authorization.k8s.io monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterrolebindings.rbac.authorization.k8s.io monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterrolebindings.rbac.authorization.k8s.io monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete Serviceaccounts
kubectl delete -n kyma-system serviceaccounts monitoring-alertmanager --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-auth-proxy-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system serviceaccounts monitoring-prometheus-node-exporter --ignore-not-found --timeout=$DELETE_TIMEOUT


# Delete Clusterroles
kubectl delete clusterroles.rbac.authorization.k8s.io monitoring-grafana-clusterrole --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterroles.rbac.authorization.k8s.io monitoring-kube-state-metrics --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterroles.rbac.authorization.k8s.io monitoring-operator --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterroles.rbac.authorization.k8s.io monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete clusterroles.rbac.authorization.k8s.io monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT



# Delete Istio CRs
kubectl delete -n kyma-system destinationrules.networking.istio.io monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system peerauthentications.security.istio.io monitoring-grafana-policy --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system authorizationpolicies.security.istio.io grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system virtualservices.networking.istio.io monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT


# Delete Rolebinding and roles
kubectl delete -n kyma-system rolebindings.rbac.authorization.k8s.io monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system roles.rbac.authorization.k8s.io monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete PriorityClass
kubectl delete priorityclasses.scheduling.k8s.io monitoring-priority-class --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete priorityclasses.scheduling.k8s.io monitoring-priority-class-high --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete Secrets
kubectl delete -n kyma-system secrets alertmanager-monitoring-alertmanager --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system secrets monitoring-auth-proxy-grafana-default --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system secrets monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system secrets monitoring-prometheus-scrape-confg --ignore-not-found --timeout=$DELETE_TIMEOUT

# Delete Configmaps
kubectl delete -n kyma-system configmaps eventing-dashboards-delivery --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps eventing-dashboards-jetstream --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps eventing-dashboards-latency --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps eventing-dashboards-nats-server --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps eventing-dashboards-pods --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps function-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-alertmanager-overview --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-apiserver --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-auth-proxy-grafana-templates --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-cluster-total --ignore-not-found--timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-grafana-config-dashboards --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-grafana-datasource --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-grafana-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-cluster --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-namespace --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-node --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-pod --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-workload --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-k8s-resources-workloads-namespace --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-kubelet --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-kyma-controllers-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-namespace-by-pod --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-namespace-by-workload --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-node-cluster-rsrc-use --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-node-rsrc-use --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-nodes --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-persistentvolumesusage --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-pod-total --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-prometheus --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-prometheus-istio-server --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-statefulset --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps monitoring-workload-total --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps ory-oathkeeper-maester-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps istio-control-plane-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps istio-mesh-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps istio-performance-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps istio-service-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps istio-workload-grafana-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system configmaps function-metrics-dashboard --ignore-not-found --timeout=$DELETE_TIMEOUT

# delete PVC
kubectl delete -n kyma-system persistentvolumeclaims monitoring-grafana --ignore-not-found --timeout=$DELETE_TIMEOUT
kubectl delete -n kyma-system persistentvolumeclaims prometheus-monitoring-prometheus-db-prometheus-monitoring-prometheus-0 --ignore-not-found --timeout=$DELETE_TIMEOUT

