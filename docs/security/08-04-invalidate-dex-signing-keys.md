title: Invalidate Dex signing keys
type: Tutorials
---

By default Dex in Kyma stores private and public keys used for signing and validating JWTs on a cluster using custom resources.
If for some reason private keys will leak it is required to invalidate such key pair to prevent issuing tokens and validating existing ones.
It is a critical security issue because an attacker can issue its own JWTs and call 3rd party services which have Dex JWT authentication enabled.

 
## Prerequisites
To complete this tutorial, the user must have either cluster-admin kubeconfig issued from a cloud provider or kyma kubeconfig and
[kyma-admin](03-05-roles-in-kyma.md) role assigned.

## Invalidate keys

1. Delete all singing keys on a cluster by running this command: `bash kubectl delete signingkeies.dex.coreos.com -n kyma-system --all `
   
   It is not recommended to interact with any of Dex CRs, however, in this situation it is the only way to change keys.

2. Restart dex pod : `bash kubectl delete po -n kyma-system -lapp=dex`. Dex will create new CR with key pair

3. Restart kyma system pods that are validating tokens issued from dex to force dropping existing public keys :

    `bash kubectl delete po -n kyma-system -l'app in (apiserver-proxy,iam-kubeconfig-service,console-backend-service,kiali-kcproxy,log-ui)'; kubectl delete po -n kyma-system -l 'app.kubernetes.io/name in (oathkeeper,tracing)'`

4. Restart all your applications that are validating dex JWTs internally(manually in the code) to force getting new public keys. 

>**NOTE:** Mentioned commands create downtime of dex, a couple of kyma components, and potentially your applications. 
>It is possible to use `kubectl scale` command and scale replicas, however, here downtime is used by intention. 
>It prevents issuing new tokens from dex signed by a compromised private key and to forces at least kyma applications
>to fetch new public keys, therefore reject all existing tokens signed by mentioned private key when validating JWT.