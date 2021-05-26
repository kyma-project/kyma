---
title: Set up Kiali and Grafana
type: Configuration
---

Kyma does not expose Kiali and Grafana by default. However you can still access them using port forwarding. Also, read [TBD: Expose Kyma UIs securely] to learn how to expose Kiali and Grafana securely using an identity provider of your choice.

  *TBD - input needed*

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  To access Kiali, do the following:

  1. Run the following command to forward a local port to a port on the Kiali Pod:
  ```bash
  kubectl -n kyma-system port-forward svc/kiali-server 20001:20001
  ```
  >Note: kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

  2. Open http://localhost:20001 in your browser. You shoud see Kiali UI.

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  To access Grafana, do the following:

  1. Run the following command to forward a local port to a port on the Grafana Pod:
  ```bash
  kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
  ```
  >Note: kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

  2. Open http://localhost:20001 in your browser. You should see Grafana UI.

  </details>

</div>
