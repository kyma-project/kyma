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

## Troubleshooting 

If there's an error in the process of fetching the certificates or saving them to Secrets, the CertificateRequest CR is not deleted. Instead, the controller adds the **error** section that contains a detailed error message to the CR.

To get the error message, run: 
```
kubectl get certificaterequests.applicationconnector.kyma-project.io {CERT_REQUEST_NAME} -o jsonpath={.status.error}
```
