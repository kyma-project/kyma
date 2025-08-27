# Keda Module Footprint

The Keda module consists of four workloads.
Two of them, `keda-manager` and `keda-operator`, implement the Kubernetes operator pattern and consume extra resources only when they detect changes in the resources they watch, for example, Keda CR, ScaledObject CR, etc. Usually, they are idle and consume as little as few millicores of CPU time and less than 30MB of memory. At the time of active reconciliation, the observed CPU time jumps to 5m.

Similarly to the operators, the `keda-admission-webhooks` workload stays idle most of the time and performs validation operations only when you submit a new object from `*.keda.sh` API group.

The last workload, `keda-operator-metrics-apiserver`, continuously serves metrics for the Kubernetes autoscaling components. Here, the consumption is the highest, but in the case of one or two active KEDA scalers, it stays at 5-7 millicores of CPU time.

| Name                            | CPU (cores) | Memory (bytes) |
|---------------------------------|-------------|----------------|
| keda-admission-webhooks         | 1m          | 10Mi           |
| keda-manager                    | 3m          | 23Mi           |
| keda-operator                   | 3m          | 26Mi           |
| keda-operator-metrics-apiserver | 5m          | 30Mi           |
