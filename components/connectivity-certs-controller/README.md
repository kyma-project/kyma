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
  name: certificate-request
spec:
  csrInfoUrl: "{CSR_INFO_URL_WITH_TOKEN}"
EOF
```

After a successful exchange of certificates, the controller creates new Secrets or modifies the existing ones that correspond to the updated certificates. The CertificateRequest CR is deleted.


## Connection status

When certificates are fetched successfully the `CentralConnection` CR is created.
It contains the `ManagementInfoURL`, status of a connection with central Connector Service.

The example resource looks like that:
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

The Connectivity Certs Controller check a status of the connection by calling `ManagementInfoURL` stored in `CentralConnection` CR.
If the Management Info call will fail, the `CentralConnection` will contain an error status.


## Certificate renewal

Connectivity Certs Controller renews client certificate when it is getting close to the expiration time. Renewed certificate will replace the previous one in the Secret together with new private key.

To renew the certificate immediately `spec.RenewNow` field on `CentralConnection` should be set to `true`. 


## Troubleshooting 

If there's an error in the process of fetching the certificates or saving them to Secrets, the CertificateRequest CR is not deleted. Instead, the controller adds the **error** section that contains a detailed error message to the CR.

To get the error message, run: 
```
kubectl get certificaterequests.applicationconnector.kyma-project.io {CERT_REQUEST_NAME} -o jsonpath={.status.error}
```

Similarly if the synchronization with central Connector Service or certificate renewal fails, the error status will be set on `CentralConnection` CR.

To check it, run:
```
kubectl get centralconnections.applicationconnector.kyma-project.io {CENTRAL_CONNECTION_NAME} -o jsonpath={.status.error.message}
```