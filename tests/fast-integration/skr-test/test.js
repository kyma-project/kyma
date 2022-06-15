const {
  gatherOptions,
  director,
  commerceMockTest,
  oidcE2ETest,
} = require('./index');
const {
} = require('../utils');
const {unregisterKymaFromCompass} = require('../compass');
const {deprovisionSKRInstance} = require('./provision/deprovision-skr');
const {getOrProvisionSKR} = require('./provision/provision-skr');

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
    await getOrProvisionSKR(options, shoot, skipProvisioning, provisioningTimeout);
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
