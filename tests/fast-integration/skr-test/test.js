const {
  gatherOptions,
  withSuffix,
  withInstanceID,
  provisionSKRInstance,
  director,
  commerceMockTest,
  oidcE2ETest,
} = require('./index');
const {
  getEnvOrThrow,
  genRandom,
} = require('../utils');
const {unregisterKymaFromCompass} = require('../compass');
const {deprovisionSKRInstance} = require('./provision/deprovision-skr');
const {
  getSKRConfig,
  prepareCompassResources,
  initK8sConfig,
} = require('./helpers');

const skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
const provisioningTimeout = 1000 * 60 * 30; // 30m
const deprovisioningTimeout = 1000 * 60 * 95; // 95m
let globalTimeout = 1000 * 60 * 70; // 70m
const slowTime = 5000;

describe('SKR test', function() {
  if (!skipProvisioning) {
    globalTimeout += provisioningTimeout + deprovisioningTimeout;
  }
  this.timeout(globalTimeout);
  this.slow(slowTime);

  let options = gatherOptions(); // with default values
  let shoot;
  const getShootInfoFunc = function() {
    return shoot;
  };

  before('Ensure SKR is provisioned', async function() {
    this.timeout(provisioningTimeout);
    if (skipProvisioning) {
      console.log('Gather information from externally provisioned SKR and prepare the compass resources');
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
    console.log('Preparing compass resources on the SKR instance...');
    await prepareCompassResources(shoot, options);

    console.log('Initiating K8s config...');
    await initK8sConfig(shoot);
  });


  // Run the OIDC tests
  oidcE2ETest(options, getShootInfoFunc);

  // Run the commerce mock tests
  commerceMockTest(options);

  after('Cleanup the resources', async function() {
    this.timeout(deprovisioningTimeout);
    if (!skipProvisioning) {
      await deprovisionSKRInstance(options, deprovisioningTimeout, false);
    } else {
      console.log('An external SKR cluster was used, de-provisioning skipped');
    }
    await unregisterKymaFromCompass(director, options.scenarioName);
  });
});
