---
title: Access the Application Connector on a local Kyma deployment
type: Details
---

To access the Application Connector on a local deployment of Kyma, you must add the Kyma server certificate to the trusted certificate storage of your programming environment. This is necessary to connect the external solution to your local Kyma deployment, allow client certificate exchange, and API registration.

For example, to access the Application Connector from a Java environment, run this command to add the Kyma server certificate to the Java Keystore:
```bash
sudo {JAVA_HOME}/bin/keytool -import -alias “Kyma” -keystore {JAVA_HOME}/jre/lib/security/cacerts -file {KYMA_HOME}/installation/certs/workspace/raw/server.crt
```
