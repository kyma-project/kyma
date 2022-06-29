const {
  initK8sConfig,
  getSKRConfig,
  withSuffix,
  withInstanceID,
  gatherOptions,
  keb,
  kcp,
  gardener,
} = require('../helpers');
const {getEnvOrThrow, genRandom, debug} = require('../../utils');
const {provisionSKR}= require('../../kyma-environment-broker');
const {BTPOperatorCreds} = require('../../smctl/helpers');

async function getOrProvisionSKR(options, shoot, skipProvisioning, provisioningTimeout) {
  if (skipProvisioning) {
    console.log('Gather information from externally provisioned SKR and prepare resources');
    const instanceID = getEnvOrThrow('INSTANCE_ID');
    let suffix = process.env.TEST_SUFFIX;
    if (suffix === undefined) {
      suffix = genRandom(4);
    }
    options = gatherOptions(
        withInstanceID(instanceID),
        withSuffix(suffix),
    );
    shoot = await getSKRConfig(instanceID);
  } else {
    console.log('Provisioning new SKR instance...');
    shoot = await provisionSKRInstance(options, provisioningTimeout);
  }

  console.log('Initiating K8s config...');
  await initK8sConfig(shoot);
}

async function provisionSKRInstance(options, timeout) {
  try {
    const btpOperatorCreds = BTPOperatorCreds.fromEnv();

    console.log(`\nInstanceID ${options.instanceID}`,
        `Runtime ${options.runtimeName}`, `Application ${options.appName}`, `Suffix ${options.suffix}`);

    const skr = await provisionSKR(keb,
        kcp, gardener,
        options.instanceID,
        options.runtimeName,
        null,
        btpOperatorCreds,
        options.customParams,
        timeout);

    debug('SKR is provisioned!');
    const shoot = skr.shoot;

    return shoot;
  } catch (e) {
    throw new Error(`Provisioning failed: ${e.toString()}`);
  } finally {
    debug('Fetching runtime status...');
    const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
    console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);
    await kcp.reconcileInformationLog(runtimeStatus);
  }
}


module.exports = {
  provisionSKRInstance,
  getOrProvisionSKR,
};

