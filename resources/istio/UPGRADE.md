# Istio Upgrade

This document highlights important modification in Kyma Istio implementation which shouldn't be removed when upgrading to a new version of Istio. 
To upgrade Istio to a newer version of the Istio charts in `resources/istio/charts`

## Istio 

The Istio chart should be upgraded by applying all changes introduced in the target version of Istio.

Sub-charts shouldn't contain any Kyma-specific modifications either in the sub-chart's `values.yaml` file or in the sub-chart's `yaml` files. Any necessary modifications should be introduced in the `values.yaml` file of the main Istio chart. 
Additionally, the main Istio chart has a `customization` sub-chart designed to handle all modifications that can't be applied through the `values.yaml` file of the main chart.  

The default versions of Prometheus, Grafana, and Kiali that come with Istio are not used in Kyma. Update these components without introducing any changes and make sure they remain disabled in the `requirements.yaml` file. Changes to these components should be applied in their respective standalone component charts and should be consulted with the owners of these components.

## Post-upgrade testing

After you finish upgrading Istio, provision a test cluster with a deployment of Kyma running with the version of Istio you introduced. Contact all interested parties that depend on Istio and make sure this deployment is thoroughly tested.