---
title: OAuth2 server production profile
type: Details
---

By default, every Kyma deployment is installed with the OAuth2 server using what is considered a development profile. In the case of the ORY Hydra OAuth2 server, this means that:
  - Hydra works in the "in-memory" mode and saves the registered OAuth2 clients in its memory, without the use of persistent storage.
  - Similarly to other Kyma components, it has an Istio sidecar Pod injected.

This configuration is not considered production-ready. To use the Kyma OAuth2 server in a production environment, configure Hydra to use the production profile.

## The production profile

The production profile introduces the following changes to the Hydra OAuth2 server deployment:
   - Persistence is enabled for Hydra; the registered client data is saved in an in-cluster database.
   - The Hydra OAuth2 server and the Hydra Maester controller have Istio sidecars disabled, destinationRule custom resources are created for these components.
   - A job that reads the generated database credentials and saves them to the configuration of Hydra is added.

### Persistence

The production profile for the OAuth2 server enables persistence and creates a database in which Hydra saves the registered OAuth2 clients. When you configure Hydra to use the production profile, a PostgreSQL database is installed
using the [official charts](https://github.com/helm/charts/tree/master/stable/postgresql).
The database is created in the cluster as a StatefulSet and uses a PersistentVolume that is provider-specific. This means that the PersistentVolume used by the database uses the default StorageClass of the cluster's host provider.

## Use the production profile

You can deploy a Kyma cluster with the Hydra OAuth2 server configured to use the production profile, or you can configure Hydra in a running cluster to use the production profile. Follow these steps:

<div tabs>
  <details>
  <summary>
  Install Kyma with production-ready Hydra
  </summary>

  1. Create an appropriate Kubernetes cluster for Kyma in your host environment.
  2. Apply an override that forces the Hydra OAuth2 server to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ory-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: ory
        kyma-project.io/installation: ""
    data:
      postgresql.enabled: "true"
      hydra.hydra.autoMigrate: "true"
    EOF
    ```
  3. Install Kyma on the cluster.

  </details>
  <details>
  <summary>
  Enable production profile in a running cluster
  </summary>

  >**CAUTION:** When you switch Hydra to use the production profile in a running cluster, you lose all registered clients. Using the production profile restarts the Hydra Pod, which wipes the entire "in-memory" storage used to save the registered client data by default.

  1. Apply an override that forces the Hydra OAuth2 server to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ory-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: ory
        kyma-project.io/installation: ""
    data:
      postgresql.enabled: "true"
      hydra.hydra.autoMigrate: "true"
    EOF
    ```
  2. Run the cluster [update procedure](/root/kyma/#installation-update-kyma).


  </details>

</div>
