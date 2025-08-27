# Install the SAP BTP Operator Module

<!--this content is for OS users only-->
To install the SAP BTP Operator module from the latest release, you must install BTP Manager and the SAP BTP service operator.

### Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Kubernetes cluster, or [k3d](https://k3d.io) for local installation
* [jq](https://github.com/stedolan/jq)
* `sap-btp-manager` Secret created. See [Create the `sap-btp-manager` Secret](03-00-create-btp-manager-secret.md).
  > [!NOTE]
  > If you don't create the `sap-btp-manager` Secret, the BtpOperator CR is in the `Warning` state and the message is `Secret resource not found reason: MissingSecret`.
* `kyma-system` namespace. If you don't have it in your cluster, use the following command to create it:
  
    ```bash
    kubectl create namespace kyma-system
    ```

## Procedure

1. To install BTP Manager, use the following command:

    ```bash
    kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-manager.yaml
    ```
    > **TIP:** Use the same command to upgrade the module to the latest version.

<br>

 2. To install the SAP BTP service operator, apply the sample BtpOperator CR:

    ```bash
    kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator-default-cr.yaml
    ```

3. To check the BtpOperator CR status, run the following command:

   ```sh
   kubectl get btpoperators btpoperator -n kyma-system
   ```

For more installation options, see [Install and Uninstall BTP Manager](../contributor/01-10-installation.md).
