---
title: Disable TLS/SSL certificate verification for communication with external solution
---

You can disable the [SSL certificate verification](../../01-overview/main-areas/application-connectivity/ac-04-security.md#ssl-certificate-verification) in the communication between Kyma and external solution represented by an application. This allows Kyma to send requests and data to an unsecured application without verifying its presented TLS/SSL certificate. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the TLS/SSL certificate verification is enabled when sending data and requests to every application.

Follow these steps to disable SSL certificate verification for communication between Kyma and an external application:

1. Edit the `{APPLICATION_CR_NAME}` application CR. Run:

   ```bash
   kubectl -n kyma-integration edit application {APPLICATION_CR_NAME}
   ```

2. Edit the Application in Vim. Select `i` to start editing.
3. Find the **skipVerify** parameter and change its value to `true`.
4. Select `esc`, type `:wq`, and select `enter` to save the changes and quit.
