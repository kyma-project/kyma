const uuid = require("uuid");
const { 
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
} = require("../kyma-environment-broker");
const {
  GardenerConfig,
  GardenerClient
} = require("../gardener");
const {
  debug,
  genRandom,
  initializeK8sClient,
} = require("../utils");
const t = require("./test-helpers");

describe("SKR SVCAT migration test", function() {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());
  const smAdminCreds = t.SMCreds.fromEnv();

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const runtimeID = uuid.v4();
  
  const svcatPlatform = `svcat-${suffix}`
  const btpOperatorInstance = `btp-operator-${suffix}`
  const btpOperatorBinding = `btp-operator-binding-${suffix}`

  debug(`RuntimeID ${runtimeID}`, `Runtime ${runtimeName}`, `Application ${appName}`, `Suffix ${suffix}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);  

  let skr;
  it(`Should provision SKR`, async function() {
    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName);
  });
  it(`Should save kubeconfig`, async function() {
    t.saveKubeconfig(skr.shoot.kubeconfig);
  });
  it(`Should initialize K8s`, async function() {
    await initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });
  let btpOperatorCreds;
  it(`Should instantiate SM Instance and Binding`, async function() {
    btpOperatorCreds = await t.smInstanceBinding(smAdminCreds, svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  });
  it(`Should install BTP Operator helm chart`, async function() {
    await t.installBTPOperatorHelmChart(btpOperatorCreds);
  });
  it(`Should deprovision SKR`, async function() {
    await deprovisionSKR(keb, runtimeID);
  });
  it(`Should cleanup SM instances and bindings`, async function() {
    await t.cleanupInstanceBinding(smAdminCreds, svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  });
});
