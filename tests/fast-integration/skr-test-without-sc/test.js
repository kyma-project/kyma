const {gatherOptions} = require('./');
const {provisionSKR, deprovisionSKR, KEBClient, KEBConfig} = require('../kyma-environment-broker');
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require('../compass');
const {oidcE2ETest, commerceMockTest} = require('./skr-test');
const {KCPWrapper, KCPConfig} = require('../kcp/client');
const {keb, director} = require('./helpers');
const {initializeK8sClient, debug} = require('../utils');
const {
  GardenerConfig,
  GardenerClient,
} = require('../gardener');
const {
  genRandom,
} = require('../utils');
const s = require('../smctl/helpers');

const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe('Execute SKR test', function() {
  debug.enabled = true;

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 30; // 30m
  const deprovisioningTimeout = 1000 * 60 * 95; // 95m
  before('Provision SKR', async function() {
    try {
      this.options = gatherOptions();

      const keb = new KEBClient(KEBConfig.fromEnv());
      const gardener = new GardenerClient(GardenerConfig.fromEnv());
      const smAdminCreds = s.SMCreds.fromEnv();


      const suffix = genRandom(4);
      const runtimeName = `kyma-${suffix}`;
      this.options.appName = `app-${suffix}`;

      const btpOperatorInstance = `btp-operator-${suffix}`;
      const btpOperatorBinding = `btp-operator-binding-${suffix}`;

      const btpOperatorCreds = await s.smInstanceBinding(smAdminCreds, btpOperatorInstance, btpOperatorBinding);

      console.log(`\nInstanceID ${this.options.instanceID}`,
          `Runtime ${runtimeName}`, `Application ${this.options.appName}`, `Suffix ${suffix}`);

      const skr = await provisionSKR(keb,
          kcp, gardener,
          this.options.instanceID,
          runtimeName,
          null,
          btpOperatorCreds,
          null,
          provisioningTimeout);
      this.shoot = skr.shoot;

      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);

      await addScenarioInCompass(director, this.options.scenarioName);
      await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
      initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      await kcp.reconcileInformationLog(runtimeStatus);
    }
  });

  oidcE2ETest();
  commerceMockTest();

  after('Deprovision SKR', async function() {
    try {
      await deprovisionSKR(keb, kcp, this.options.instanceID, deprovisioningTimeout);
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after deprovisioning: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    }
    await unregisterKymaFromCompass(director, this.options.scenarioName);
  });
});
