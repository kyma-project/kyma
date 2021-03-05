const {
  KEBClient,
  KEBConfig,
  provisionSKR,
  ensureOperationSucceeded,
  getShootName,
} = require("../keb");

const { initializeK8sClient, k8sCoreV1Api, debug, kc, fromBase64 } = require("../utils");

const fs = require("fs");

describe("Provisioning SKR", function () {
  let kebConfig = new KEBConfig();

  
  const kebClient = new KEBClient(kebConfig);

  it("Send SKR provisioning call to KEB", async function () {
  const  {operationID, shootName} = await provisionSKR(
      kebClient,
      kebConfig.instanceID,
      kebConfig.planID,
      kebConfig.name
    );
 
    debug("provision called");
    debug(operationID);
    debug(shootName);

    await ensureOperationSucceeded(
      kebClient,
      kebConfig.instanceID,
      operationID
    );

    debug("provisioned");

    initializeK8sClient();
    let secretName = `${shootName}.kubeconfig`;
    debug(kc.getContexts());

    let secret = await k8sCoreV1Api.readNamespacedSecret(
      secretName,
      "garden-kyma-dev"
    );

    let b64kubeconfig = secret.body.data["kubeconfig"];
    let strKubeconfig = fromBase64(b64kubeconfig)

    fs.writeFile("~/.kube/config", strKubeconfig, () => {
      initializeK8sClient();
    });
  }).timeout(70000);
});
