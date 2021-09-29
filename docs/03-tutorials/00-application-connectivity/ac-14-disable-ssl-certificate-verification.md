---
title: Disable SSL certificate verification
---

<!-- TODO 9: describe how to do that in the Central Gateway scenario -->
You can disable the [SSL certificate verification](../../01-overview/main-areas/application-connectivity/ac-04-security.md#ssl-certificate-verification) in the communication between Kyma and an application to allow Kyma to send requests and data to an unsecured application. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the SSL certificate verification is enabled when sending data and requests to every application.

Follow these steps to disable SSL certificate verification for communication between Kyma and an external application:

1. Edit the `{APPLICATION_CR_NAME}-application-gateway` Deployment in the `kyma-integration` Namespace. Run:

   ```bash
   kubectl -n kyma-integration edit deployment {APPLICATION_CR_NAME}-application-gateway
   ```

2. Edit the Deployment in Vim. Select `i` to start editing.
3. Find the **skipVerify** parameter and change its value to `true`.
4. Select `esc`, type `:wq`, and select `enter` to save the changes and quit.
