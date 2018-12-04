---
title: Installation with Knative
type: Installation
---

You can install Kyma with [Knative](https://cloud.google.com/knative/) and use its solutions for handling events and serverless functions.

## Knative with local deployment from sources

When you install Kyma locally from sources, add the `--knative` argument to the `run.sh` script. Run:

```
./run.sh --knative
```

## Knative with a GKE cluster deployment from sources

To install Kyma with Knative when deploying on a GKE cluster from sources, follow the instructions outlined in the **Install Kyma on a GKE cluster** installation guide.

To prepare the `my-kyma.yaml` configuration file that installs Kyma with Knative on a GKE cluster, run:

```
cat installation/resources/installer.yaml <(echo -e "\n---") installation/resources/installer-config-cluster.yaml.tpl  <(echo -e "\n---") installation/resources/installer-cr-cluster.yaml.tpl | sed -e "s/__KNATIVE__/true/g | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```
