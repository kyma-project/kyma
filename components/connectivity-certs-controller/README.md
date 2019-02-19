# Connectivity Certs Controller


## Overview

Connectivity Certs Controller is responsible for communicating with the Connector Service.
It fetches both a client certificate and the root CA certificate and saves them to secrets.


## Fetching certificates

Certificates Manager is operator which reacts to `CertificateRequest` custom resource. 
It requires `csrInfoUrl` field.
To create it run:
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

After successful exchange the secrets will be created or modified and the `CertificateRequest` will be deleted.  


## Troubleshooting 

If something went wrong while requesting certificates or saving them to secrets, the custom resource will not be deleted and it will have and `error` status with a detailed error message.
To get the error message run: 
```
kubectl get certificaterequests.applicationconnector.kyma-project.io {CERT_REQUEST_NAME} -o jsonpath={.status.error}
```
