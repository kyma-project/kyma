# Create the `sap-btp-manager` Secret

<!--this content is for OS users only-->
To create the `sap-btp-manager` Secret, follow these steps:
1. To obtain the access credentials to the service instance, create a service binding.  For details, see [Installation and Setup: Obtain the access credentials for the SAP BTP service operator](https://github.com/SAP/sap-btp-service-operator/blob/main/README.md#installation-and-setup).
2. Copy and save the access credentials into your `creds.json` file in your working directory. 
3. In the same directory, run the following script to create the Secret:
   
   ```sh
   curl https://raw.githubusercontent.com/kyma-project/btp-manager/main/hack/create-secret-file.sh | bash -s
   ```

    The expected result is the file `operator-secret.yaml` created in your working directory:

    ```yaml
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: sap-btp-manager
      namespace: kyma-system
      labels:
        app.kubernetes.io/managed-by: kcp-kyma-environment-broker
    data:
      clientid: {CLIENT_ID}
      clientsecret: {CLIENT_SECRET}
      sm_url: {SM_URL}
      tokenurl: {AUTH_URL}
      cluster_id: {CLUSTER_ID}
    ```
4. To create the Secret, run:

    ```
    kubectl create -f ./operator-secret.yaml
    ```

    You see the status `secret/sap-btp-manager created`.