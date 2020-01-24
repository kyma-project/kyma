# knative-eventing-init

This helm chart sets the default knative eventing channel. By default, Natss will be used, but a different channel can be configured in `values.yaml`.

This chart has to be installed before `knative-eventing`, because the [knative eventing-webhook](https://github.com/knative/eventing/tree/master/cmd/webhook) requires the default channel config map to exist. Otherwise it will refuse to start.

## Install chart

```bash
cd resources
helm install -n knative-eventing-init ./knative-eventing-init
```
