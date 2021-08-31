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
  switchDebug,
} = require("../utils");
const t = require("./test-helpers");
const sampleResources = require("./deploy-sample-resources");

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
  switchDebug(on = true)
  debug(`RuntimeID ${runtimeID}`, `Runtime ${runtimeName}`, `Application ${appName}`, `Suffix ${suffix}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  let platformCreds;
  it(`Should Provision Platform`, async function() {
    platformCreds = await t.provisionPlatform(smAdminCreds, svcatPlatform)
  });

  let btpOperatorCreds;
  it(`Should instantiate SM Instance and Binding`, async function() {
    btpOperatorCreds = await t.smInstanceBinding(btpOperatorInstance, btpOperatorBinding);
  });

  let skr;
  it(`Should provision SKR`, async function() {
    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName, platformCreds, btpOperatorCreds);
  });

  it(`Should save kubeconfig`, async function() {
    t.saveKubeconfig(skr.shoot.kubeconfig);
  });

  it(`Should initialize K8s`, async function() {
    await initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });

  let clusterid
  it('Should read cluster id from Service Catalog', async function() {
    clusterid = await  t.readClusterID()
    debug('Found Service Catalog ClusterID: ' + clusterid)
  })

  it(`Should install sample service catalogue resources`, async function() {
    await sampleResources.deploy()
  });

  it('Should mark the Platform for migration', async function() {
    await t.markForMigration(smAdminCreds, platformCreds.clusterId, btpOperatorCreds.instanceId)
  })

  it(`Should install BTP Operator helm chart`, async function() {
    await t.installBTPOperatorHelmChart(btpOperatorCreds, clusterid);
  });

  let secretsAndPresets
  it(`Should store secrets and presets`, async function() {
    secretsAndPresets = await sampleResources.storeSecretsAndPresets()
  });

  it(`Should install BTP Service Operator Migration helm chart`, async function() {
    await t.installBTPServiceOperatorMigrationHelmChart();

    // TODO: Print log output of migrator job "sap-btp-operator-migration"
  });

  // // TODO: Remove
  // // this sleep is created to have a time to check the cluster before deprovisioning it
  // it(`Should Sleep and wakeup properly`, async function() {
  //   await sampleResources.goodNight()
  // });
  //
  // it(`Should pass sanity check`, async function() {
  //   // TODO: Wait/Check until Job of BTP-Migrator/SC-Removal is finished successfully
  //   let existing = await sampleResources.storeSecretsAndPresets()
  //   // Check if Secrets and PodPresets are still available
  //   await sampleResources.checkSecrets(existing.secrets)
  //   await sampleResources.checkPodPresets(secretsAndPresets.podPresets, existing.podPresets)
  //
  //   // TODO: Check if all other SVCat resources are successfully removed
  // });
  //
  //
  // it(`Should destroy sample service catalogue ressources`, async function() {
  //   // TODO: Remove anything from BT-Operator
  //   await sampleResources.destroy()
  //
  //   // TODO: Check if no Service Instances are left over
  // });
  //
  // it(`Should deprovision SKR`, async function() {
  //   await deprovisionSKR(keb, runtimeID);
  // });
  //
  // it(`Should cleanup platform --cascade, operator instances and bindings`, async function() {
  //   await t.cleanupInstanceBinding(smAdminCreds, svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  // });
});
