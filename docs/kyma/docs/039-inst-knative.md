---
title: Installation with Knative
type: Installation
---

You can install Kyma with [Knative](https://cloud.google.com/knative/) and use its solutions for handling events and serverless functions.

> **NOTE:** You can’t install Kyma with Knative on clusters with a pre-allocated ingress gateway IP address.

> **NOTE:** Knative intagration requires Kyma 0.6 or higher.

## Knative with local deployment from release

When you install Kyma locally from a release, follow [this](#installation-install-kyma-locally-from-the-release-install-kyma-on-minikube) guide and run the following command after you complete step 6:
```
kubectl -n kyma-installer patch configmap installation-config-overrides -p '{"data": {"knative": "true"}}'
```  

## Knative with local deployment from sources

When you install Kyma locally from sources, add the `--knative` argument to the `run.sh` script. Run this command:

```
./run.sh --knative
```

## Knative with a GKE cluster deployment from release

To install Kyma with Knative when deploying on a GKE cluster from release, follow the instructions outlined in the **Install Kyma on a GKE cluster** installation guide.

To prepare the `my-kyma.yaml` configuration file that installs Kyma with Knative on a GKE cluster, run:

```
cat kyma-config-cluster.yaml | sed -e "s/__DOMAIN__/$DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/global.knative:.*/global.knative: \"true\"/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g"  >my-kyma.yaml
```

## Knative with a GKE cluster deployment from sources

To install Kyma with Knative when deploying on a GKE cluster from sources, follow the instructions outlined in the **Install Kyma on a GKE cluster** installation guide.

To prepare the `my-kyma.yaml` configuration file that installs Kyma with Knative on a GKE cluster, run:

```
cat installation/resources/installer.yaml <(echo -e "\n---") installation/resources/installer-config-cluster.yaml.tpl  <(echo -e "\n---") installation/resources/installer-cr-cluster.yaml.tpl | sed -e "s/global.knative:.*/global.knative: \"true\"/g" | sed -e "s/__DOMAIN__/$DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```
