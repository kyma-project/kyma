const {
  keb,
  gatherOptions,
  withCustomParams,
} = require('../skr-test/helpers');
const {
  debug,
  getEnvOrThrow,
  switchDebug,
} = require('../utils');
const {getOrProvisionSKR} = require('../skr-test/provision/provision-skr');
const {deprovisionAndUnregisterSKR} = require('../skr-test/provision/deprovision-skr');
const {upgradeSKRInstance} = require('./upgrade/upgrade-skr');
const {
  commerceMockTestPreparation,
  commerceMockTests,
  commerceMockCleanup,
} = require('../skr-test');

const skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
const provisioningTimeout = 1000 * 60 * 60; // 1h
const deprovisioningTimeout = 1000 * 60 * 30; // 30m
const upgradeTimeoutMin = 30; // 30m
let globalTimeout = 1000 * 60 * 90; // 90m
const slowTime = 5000;

const kymaVersion = getEnvOrThrow('KYMA_VERSION');
const kymaUpgradeVersion = getEnvOrThrow('KYMA_UPGRADE_VERSION');

describe('SKR-Upgrade-test', function() {
  switchDebug(on = true);

  if (!skipProvisioning) {
    globalTimeout += provisioningTimeout + deprovisioningTimeout; // 3h
  }
  this.timeout(globalTimeout);
  this.slow(slowTime);

  const customParams = {
    'kymaVersion': kymaVersion,
  };

  let options = gatherOptions(
      withCustomParams(customParams),
  );
  let skr;

  debug(
      `PlanID ${getEnvOrThrow('KEB_PLAN_ID')}`,
      `SubAccountID ${keb.subaccountID}`,
      `InstanceID ${options.instanceID}`,
      `Scenario ${options.scenarioName}`,
      `Runtime ${options.runtimeName}`,
      `Application ${options.appName}`,
  );

  // Credentials for KCP OIDC Login

  // process.env.KCP_TECH_USER_LOGIN    =
  // process.env.KCP_TECH_USER_PASSWORD =
  process.env.KCP_OIDC_ISSUER_URL = 'https://kymatest.accounts400.ondemand.com';
  // process.env.KCP_OIDC_CLIENT_ID     =
  // process.env.KCP_OIDC_CLIENT_SECRET =
  process.env.KCP_KEB_API_URL = 'https://kyma-env-broker.cp.dev.kyma.cloud.sap';
  process.env.KCP_GARDENER_NAMESPACE = 'garden-kyma-dev';
  process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
  process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';


  // SKR Provisioning
  before(`Provision SKR with ID ${options.instanceID} and version ${kymaVersion}`, async function() {
    this.timeout(provisioningTimeout);
    await getOrProvisionSKR(options, skr, skipProvisioning, provisioningTimeout);
  });

  // Perform Tests before Upgrade
  it('Execute Commerce Mock Tests', async function() {
    commerceMockTestPreparation(options);
    commerceMockTests(options.testNS);
  });

  // Upgrade
  it('Perform Upgrade', async function() {
    await upgradeSKRInstance(options, kymaUpgradeVersion, upgradeTimeoutMin);
  });

  // Perform Tests after Upgrade
  it('Execute commerceMockTests', async function() {
    commerceMockTests(options.testNS);
  });

  // Cleanup
  const skipCleanup = getEnvOrThrow('SKIP_CLEANUP');
  if (skipCleanup === 'FALSE') {
    after('Cleanup the resources', async function() {
      this.timeout(deprovisioningTimeout);
      await commerceMockCleanup(options.testNS);
      await deprovisionAndUnregisterSKR(deprovisioningTimeout, skipProvisioning);
    });
  }
});
