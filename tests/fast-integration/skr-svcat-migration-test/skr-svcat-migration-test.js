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
  waitForJob,
  printContainerLogs,
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
  it(`Should provision new ServiceManager platform`, async function() {
    platformCreds = await t.provisionPlatform(smAdminCreds, svcatPlatform)
  });

  let btpOperatorCreds;
  it(`Should instantiate ServiceManager instance and binding for BTP operator`, async function() {
    btpOperatorCreds = await t.smInstanceBinding(btpOperatorInstance, btpOperatorBinding);
  });

  let skr;
  it(`Should provision SKR`, async function() {
    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName, platformCreds, btpOperatorCreds);
  });

  it(`Should save kubeconfig for the SKR to ~/.kube/config`, async function() {
    t.saveKubeconfig(skr.shoot.kubeconfig);
  });

  it(`Should initialize K8s client`, async function() {
    await initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });

  let clusterid
  it('Should read cluster id from Service Catalog ConfigMap', async function() {
    clusterid = await  t.readClusterID()
    debug('Found Service Catalog ClusterID: ' + clusterid)
  })

  it(`Should install sample Service Catalog resources`, async function() {
    await sampleResources.deploy()
  });

  it('Should mark the platform for migration in Service Manager', async function() {
    await t.markForMigration(smAdminCreds, platformCreds.clusterId, btpOperatorCreds.instanceId)
  })

  it(`Should install BTP operator helm chart`, async function() {
    await t.installBTPOperatorHelmChart(btpOperatorCreds, clusterid);
  });

  let secretsAndPresets
  it(`Should store secrets and presets of sample resources`, async function() {
    secretsAndPresets = await sampleResources.storeSecretsAndPresets()
  });

  it(`Should check if pod presets injected secrets to functions containers`, async function() {
    await t.checkPodPresetEnvInjected();
  });

  it(`Should install BTP service operator migration helm chart`, async function() {
    await t.installBTPServiceOperatorMigrationHelmChart();
  });

  it(`Should wait for migration job to finish`, async function() {
    await waitForJob("sap-btp-operator-migration", "sap-btp-operator");
  });
  
  it(`Should print the container logs of the migration job`, async function() {
    await printContainerLogs('job-name=sap-btp-operator-migration', 'migration', 'sap-btp-operator');
  });

  it(`Should still contain pod presets and the secrets`, async function() {
    let existing = await sampleResources.storeSecretsAndPresets()
    // Check if Secrets and PodPresets are still available
    await sampleResources.checkSecrets(existing.secrets)
    await sampleResources.checkPodPresets(secretsAndPresets.podPresets, existing.podPresets)
  });

  it(`Should restart functions pods`, async function() {
    await t.restartFunctionsPods();
  });

  it(`Should check if pod presets injected secrets in functions containers are present after migration`, async function() {
    await t.checkPodPresetEnvInjected();
  });

  it(`Should destroy sample service catalog resources`, async function() {
    // TODO: Remove anything from BT-Operator
    await sampleResources.destroy()

    // TODO: Check if no Service Instances are left over
  });

  it(`Should deprovision SKR`, async function() {
    await deprovisionSKR(keb, runtimeID);
  });

  it(`Should cleanup platform --cascade, operator instances and bindings`, async function() {
    await t.cleanupInstanceBinding(smAdminCreds, svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  });
});
