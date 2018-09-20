```
  __  __             _ _             _
 |  \/  |           (_) |           (_)
 | \  / | ___  _ __  _| |_ ___  _ __ _ _ __   __ _
 | |\/| |/ _ \| '_ \| | __/ _ \| '__| | '_ \ / _` |
 | |  | | (_) | | | | | || (_) | |  | | | | | (_| |
 |_|  |_|\___/|_| |_|_|\__\___/|_|  |_|_| |_|\__, |
                                              __/ |
                                             |___/
```

## Overview

The [Kube-Prometheus](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus) implementation provides end-to-end Kubernetes cluster monitoring in [Kyma](https://github.com/kyma-project/kyma) using the [Prometheus operator](https://github.com/coreos/prometheus-operator).

This chart installs [Prometheus](https://prometheus.io/), [Alertmanager](https://github.com/prometheus/alertmanager), and [Grafana](https://grafana.com/), along with the configuration to monitor a Kubernetes cluster. It requires a running instance of Prometheus operator, which Kyma provides.

`kube-prometheus` installs in the `kyma-system` Namespace.

## Details

* [Grafana in Kyma](charts/grafana/README.md)
