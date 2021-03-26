---
title: Invalidate Dex signing keys
type: Tutorials
---

By default, Dex in Kyma stores private and public keys used to sign and validate JWT tokens on a cluster using custom resources. If, for some reason, the private keys leak, you must invalidate the private-public key pair to prevent the attacker from issuing tokens and validating the existing ones.
It is critical to do so, because otherwise the attackers can use a private key to issue a new JWT token to call third party services which have Dex JWT authentication enabled. 
Follow this tutorial to learn how to invalidate the signing keys.

## Prerequisites

To complete this tutorial, you must have either the cluster-admin `kubeconfig` file issued from a cloud provider or the Kyma `kubeconfig` file the and [**kyma-admin**](#details-roles-in-kyma) role assigned.

## Steps

Perform these steps to invalidate the keys: 

1. Delete all signing keys on a cluster:

```bash
kubectl delete signingkeies.dex.coreos.com -n kyma-system --all 
```

>**NOTE:** Although it is not recommended to interact with any of Dex CRs, in this situation it is the only way to invalidate the keys.

2. Restart the Dex Pod:

```bash
kubectl delete po -n kyma-system -lapp=dex
```
Dex will create a new CR with a private-public key pair.

3. Restart `kyma-system` Pods that validate tokens issued from Dex to drop the existing public keys:

```bash 
kubectl delete po -n kyma-system -l'app in (apiserver-proxy,iam-kubeconfig-service,console-backend-service,kiali-kcproxy,log-ui)'; kubectl delete po -n kyma-system -l 'app.kubernetes.io/name in (oathkeeper,tracing)'
```

4. Manually restart all your applications that validate Dex JWT tokens internally to get the new public keys. 

>**NOTE:** Following the tutorial steps results in the downtime of Dex, a couple of Kyma components, and potentially your applications. If you want to use `kubectl scale` command to scale replicas, bear in mind that this downtime is intentional. It prevents Dex from issuing new tokens signed by a compromised private key and forces at least Kyma applications to fetch new public keys, and at the same time reject all existing tokens signed by the compromised private key during JWT token validation.
