---
title: Disable TLS certificate verification
---

You can disable the [TLS certificate verification](../../01-overview/main-areas/application-connectivity/ac-04-security.md#tls-certificate-verification-for-external-systems) for the connections between Kyma and an external solution represented by an Application. This allows Kyma to send requests and data to an unsecured Application without verifying its presented TLS certificate. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the TLS certificate verification is enabled when sending data and requests to every Application.
Follow these steps to disable TLS certificate verification:

1. Edit the `{APP_NAME}` Application CR. Run:

   ```bash
   kubectl edit application {APP_NAME}
   ```

2. Edit the Application by setting the **skipVerify** parameter to `true`.
3. Save the changes and quit.