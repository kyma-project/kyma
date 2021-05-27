---
title: Set up Kiali and Grafana
type: Configuration
---
<<<<<<< HEAD

Kyma does not expose Kiali and Grafana by default. However you can still access them using port forwarding. Also, read [Expose Kyma UIs securely](http://tbd) to learn how to expose Kiali and Grafana securely using an identity provider of your choice.

## Prerequisites

- You have defined the kubeconfig file for your cluster as default (see [Kubernetes: Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)).
=======
Learn how to configure Kiali and Grafana in regards to port forwarding, securing and exposing.

  *TBD - input needed*
>>>>>>> kyma-2.0-docu

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

<<<<<<< HEAD
  To access Kiali, do the following:

  1. Run the following command to forward a local port to a port on the Kiali Pod:
  ```bash
  kubectl -n kyma-system port-forward svc/kiali-server 20001:20001
  ```
  >**NOTE:** kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

  2. Open http://localhost:20001 in your browser. You shoud see Kiali UI.
=======
  To configure Kiali, to the following:

  1. 
  1. 
  1. 

>>>>>>> kyma-2.0-docu

  </details>
  <details>
  <summary>
  Grafana
  </summary>

<<<<<<< HEAD
  To access Grafana, do the following:

  1. Run the following command to forward a local port to a port on the Grafana Pod:
  ```bash
  kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
  ```
  >**NOTE:** kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

  2. Open http://localhost:3000 in your browser. You should see Grafana UI.

  </details>

</div>
=======
  To configure Grafana, do the following:

  1.
  2.
  3.

  </details>

</div>
>>>>>>> kyma-2.0-docu
