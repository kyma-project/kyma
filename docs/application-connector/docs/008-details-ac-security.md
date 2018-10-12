---
title: Security
type: Details
---

## Client certificates

To provide maximum security, the Application Connector uses TLS protocol with Client Authentication enabled. As a result, whoever wants to connect to the Application Connector must present a valid client certificate, which is dedicated to a specific Remote Environment. In this way, the traffic is fully encrypted and the client has a valid identity.

## Disable SSL certificate verification

You can disable the SSL certificate verification in the communication between Kyma and a Remote Environment to allow Kyma to send requests and data to an unsecured Remote Environment. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the SSL certificate verification is enabled when sending data and requests to every Remote Environment.

Follow these steps to disable SSL certificate verification for communication between Kyma and an existing Remote Environment:

  1. Edit the `ec-default-proxy-service` Deployment in the `kyma-integration` Namespace. Run:
    ```
    kubectl -n kyma-integration edit deployment ec-default-proxy-service
    ```
  2. Edit the Deployment in Vim. Select `i` to start editing.
  3. Find the **skipVerify** parameter and change its value to `true`.
  4. Select `esc`, type `:wq`, and select `enter` to write and quit.
