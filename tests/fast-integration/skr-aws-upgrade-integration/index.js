const {
  gatherOptions,
  commerceMockTest,
  withCustomParams,
} = require('../skr-test');
const {
  getEnvOrThrow,
  switchDebug,
} = require('../utils');
const {getOrProvisionSKR} = require('../skr-test/provision/provision-skr');
const {deprovisionAndUnregisterSKR} = require('../skr-test/provision/deprovision-skr');
const {upgradeSKRInstance} = require('./upgrade/upgrade-skr');

const skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
const provisioningTimeout = 1000 * 60 * 60; // 1h
const deprovisioningTimeout = 1000 * 60 * 30; // 30m
const upgradeTimeoutMin = 30; // 30m
let globalTimeout = 1000 * 60 * 90; // 90m
const slowTime = 5000;

const kymaVersion = getEnvOrThrow('KYMA_VERSION');
const kymaUpgradeVersion = getEnvOrThrow('KYMA_UPGRADE_VERSION');

describe('SKR-Upgrade-test', function() {
  switchDebug(true);

  if (!skipProvisioning) {
    globalTimeout += provisioningTimeout + deprovisioningTimeout; // 3h
  }
  this.timeout(globalTimeout);
  this.slow(slowTime);

  const customParams = {
    'kymaVersion': kymaVersion,
  };

  const options = gatherOptions(
      withCustomParams(customParams),
  );
  let shoot;

  before(`Provision SKR with ID ${options.instanceID} and version ${kymaVersion}`, async function() {
    this.timeout(provisioningTimeout);
    await getOrProvisionSKR(options, shoot, skipProvisioning, provisioningTimeout);
  });

  it('Execute Commerce Mock Tests', async function() {
    commerceMockTest(options);
  });

  it('Perform Upgrade', async function() {
    await upgradeSKRInstance(options, kymaUpgradeVersion, upgradeTimeoutMin);
  });

  it('Execute commerceMockTests', async function() {
    commerceMockTest(options);
  });

  const skipCleanup = getEnvOrThrow('SKIP_CLEANUP');
  if (skipCleanup === 'FALSE') {
    after('Cleanup the resources', async function() {
      this.timeout(deprovisioningTimeout);
      await deprovisionAndUnregisterSKR(options, deprovisioningTimeout, skipProvisioning);
    });
  }
});
