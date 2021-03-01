const {
  KEBClient,
  KEBConfig,
  provisionSKR,
  ensureOperationSucceeded,
  getProvisionerID,
} = require("../keb");

const {
  ProvisionerClient,
  ProvisionerConfig,
  getKubeconfig,
} = require("../provisioner");

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
    const provisionerID = await getProvisionerID(kebClient, config.instanceID);
    await getKubeconfig(provisionerClient, provisionerID);
  });
});
