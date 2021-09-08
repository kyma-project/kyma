const uuid = require("uuid");
const { 
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
} = require("../kyma-environment-broker");
const {
  GardenerConfig,
  GardenerClient,
} = require("../gardener");
const {
  debug,
} = require("../utils");
const {
  help,
} = require("helpers")

describe("SKR test", function() {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);  
  
  it(`Provision SKR with ID ${runtimeID}`, async function() {
    await help()
    // skr = await provisionSKR(keb, gardener, runtimeID, runtimeName);
    // initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });

});