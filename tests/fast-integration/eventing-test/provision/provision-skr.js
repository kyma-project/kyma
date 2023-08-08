const {
  KEBConfig,
  KEBClient,
  provisionSKR,
} = require('../kyma-environment-broker');
const {GardenerConfig, GardenerClient} = require('../../gardener');
const {
  debug,
  getEnvOrThrow,
} = require('../../utils');
const {
  KCPConfig,
  KCPWrapper,
} = require('../kcp/client');
const {
  saveKubeconfig,
} = require('../skr-helpers/helpers');
const {slowTime} = require('../utils');

describe('SKR eventing test', function() {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(slowTime);
  const provisioningTimeout = 1000 * 60 * 60; // 1h

  // initialize the clients required for skr provisioning
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());
  const kcp = new KCPWrapper(KCPConfig.fromEnv());

  const kymaVersion = getEnvOrThrow('KYMA_VERSION');
  const runtimeName = getEnvOrThrow('RUNTIME_NAME');
  const instanceId = getEnvOrThrow('INSTANCE_ID');
  const subAccountID = keb.subaccountID;
  const kymaOverridesVersion = process.env.KYMA_OVERRIDES_VERSION || '';

  let skr;

  debug(
      `PlanID ${getEnvOrThrow('KEB_PLAN_ID')}`,
      `SubAccountID ${subAccountID}`,
      `instanceId ${instanceId}`,
      `Runtime ${runtimeName}`,
  );

  const customParams = {'kymaVersion': kymaVersion};
  if (kymaOverridesVersion) {
    customParams['overridesVersion'] = kymaOverridesVersion;
  }

  // SKR Provisioning
  it(`Perform kcp login`, async function() {
    const version = await kcp.version([]);
    debug(version);

    await kcp.login();
  });

  it(`Provision SKR with ID ${instanceId}`, async function() {
    console.log(`Provisioning SKR with version ${kymaVersion}`);
    debug(`Provision SKR with Custom Parameters ${JSON.stringify(customParams)}`);
    skr = await provisionSKR(
        keb,
        kcp,
        gardener,
        instanceId,
        runtimeName,
        null,
        null,
        customParams,
        provisioningTimeout);
  });

  it(`Should get Runtime Status after provisioning`, async function() {
    const runtimeStatus = await kcp.getRuntimeStatusOperations(instanceId);
    console.log(`\nRuntime status: ${runtimeStatus}`);
    await kcp.reconcileInformationLog(runtimeStatus);
  });

  it(`Should save kubeconfig for the SKR to ~/.kube/config`, async function() {
    await saveKubeconfig(skr.shoot.kubeconfig);
  });
});
