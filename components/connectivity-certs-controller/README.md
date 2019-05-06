# Connectivity Certs Controller

## Overview

The Connectivity Certs Controller fetches the client certificate and the root CA from the central Connector Service and saves them to Secrets.

## Fetching certificates

The Controller acts on to CertificateRequest custom resource (CR). It requires the `csrInfoUrl` field.

To create the CR, run:
```
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: CertificateRequest
metadata:
  name: {CERTIFICATE_REQUEST_NAME}
spec:
  csrInfoUrl: "{CSR_INFO_URL_WITH_TOKEN}"
EOF
```

After a successful exchange of certificates, the controller creates new Secrets or modifies the existing ones that correspond to the updated certificates. The CertificateRequest CR is deleted.


## Connection status

When certificates are fetched successfully, the CentralConnection CR is created.
It contains **ManagementInfoURL**, the status of a connection with the central Connector Service, and the certificate validity period.


The CentralConnection CR is created with the same name as the name of `CertificateRequest` for which the connection was established.
To get the CentralConnection CR, run:
```
kubectl get centralconnections.applicationconnector.kyma-project.io {CENTRAL_CONNECTION_NAME} -oyaml
```

The example resource looks as follows:
```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: CentralConnection
metadata:
  name: my-connection
spec:
  establishedAt: "2019-04-10T08:47:38Z"
  managementInfoUrl: https://connector-service.central.io/v1/runtimes/management/info
status:
  certificateStatus:
    notAfter: "2019-07-11T08:47:38Z"
    notBefore: "2019-04-10T08:47:38Z"
  synchronizationStatus:
    lastSuccess: "2019-04-10T09:01:06Z"
    lastSync: "2019-04-10T09:01:06Z"
```

The Connectivity Certs Controller checks the connection status by calling **ManagementInfoURL** stored in the CentralConnection CR.


## Certificate renewal

The Connectivity Certs Controller renews a client certificate when it gets close to the expiration time. A renewed certificate will replace the previous one in the Secret together with a new private key.

To renew the certificate, set the **spec.renewNow** field in the CentralConnection CR to `true`. 


## Troubleshooting 

If there's an error in the process of fetching the certificates or saving them to Secrets, the CertificateRequest CR is not deleted. Instead, the controller adds the **error** section that contains a detailed error message to the CR.

To get the error message, run: 
```
kubectl get certificaterequests.applicationconnector.kyma-project.io {CERT_REQUEST_NAME} -o jsonpath={.status.error}
```

Similarly, if the synchronization with the central Connector Service or the certificate renewal fails, the error status will be displayed in the CentralConnection CR.

To check the error status, run:
```
kubectl get centralconnections.applicationconnector.kyma-project.io {CENTRAL_CONNECTION_NAME} -o jsonpath={.status.error.message}
```
