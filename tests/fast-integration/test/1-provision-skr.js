const {
  KEBClient,
  KEBConfig,
  provisionSKR,
  ensureOperationSucceeded,
  getShootName,
} = require("../keb");

const {
  k8sCoreV1Api
} = require("../utils")

const {
  fs
} = require("fs")

describe("Provisioning SKR", function () {
  let kebConfig = new KEBConfig();
  let provisionerConfig = new ProvisionerConfig();
  let operationID;

  const kebClient = new KEBClient(kebConfig);
  const provisionerClient = new ProvisionerClient(provisionerConfig);
  it("Send provisioning call to KEB", async function () {
    operationID = await provisionSKR(
      kebClient,
      config.instanceID,
      config.planID,
      config.name
    );
  });

  it("Wait for the SKR to provision", async function () {
    await ensureOperationSucceeded(kebClient, config.instanceID, operationID);
  }).timeout(3600000);

  it("Download SKR Kubeconfig", async function () {
    const shootName = await getShootName(kebClient, config.instanceID);
    let secret = await k8sCoreV1Api.readNamespacedSecret(`${shootName}.kubeconfig`);
    let b64kubeconfig = secret.body.data["kubeconfig"]
    const buff = Buffer.from(b64kubeconfig, 'base64');
    const str = buff.toString('utf-8');
    fs.writeFile("kubeconfig.yaml", str)
  });
});
