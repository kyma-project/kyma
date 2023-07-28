---
title: Disable TLS certificate verification
---

You can disable the [TLS certificate verification](../../01-overview/application-connectivity/ac-04-security.md#tls-certificate-verification-for-external-systems) for the connections between Kyma and an external solution represented by an Application. This allows Kyma to send requests and data to an unsecured Application without verifying its presented TLS certificate. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the TLS certificate verification is enabled when sending data and requests to every Application.

To disable TLS certificate verification [export your Application name as an environment variable](./ac-01-create-application.md#prerequisites) and follow these steps: 

1. Edit the Application custom resource (CR) for your Application. Run:

   ```bash
   kubectl edit application $APP_NAME
   ```

2. Edit the Application by setting the **skipVerify** parameter to `true`.
3. Save the changes and quit.