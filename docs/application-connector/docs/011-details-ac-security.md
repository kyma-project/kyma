---
title: Application Connector security
type: Details
---

To provide maximum security, the Application Connector uses TLS protocol with Client Authentication enabled. As a result, whoever wants to connect to the Application Connector must present a valid client certificate, which is dedicated to a specific Remote Environment. In this way, the traffic is fully encrypted and the client has a valid identity. Kyma representatives generate client certificates.

### Disable SSL certificate verification

You can disable the SSL certificate verification in the communication between Kyma and a Remote Environment to allow Kyma to send requests and data to an unsecured Remote Environment. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the SSL certificate verification is enabled when sending data and requests to every Remote Environment.

* Disable SSL certificate verification for communication between Kyma and an existing Remote Environment

  - Edit the `ec-default-gateway` Deployment in the `kyma-integration` Namespace. Run:
    ```
    kubectl -n kyma-integration edit deployment ec-default-gateway
    ```
  - Edit the Deployment in Vim. Select `i` to start editing.
  - Find the **skipVerify** parameter and change its value to `true`.
  - Select `esc`, type `:wq`, and select `enter` to write and quit.

* Install a new Remote Environment with the certificate verification disabled

  Disable the certification by adding the `--skipVerify=true` flag to the `helm install` command. This is an example of a command that installs a new Remote Environment with SSL certificate verification disabled:

  ```
  helm install --name {remote-environment-name} --set deployment.args.sourceType=commerce --set global.isLocalEnv=false --set global.domainName={domain-name} --namespace kyma-integration ./remote-environments --skipVerify=true  
  ```
