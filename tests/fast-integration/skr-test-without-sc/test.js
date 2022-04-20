const {gatherOptions} = require('./');
const {provisionSKR, deprovisionSKR, KEBClient, KEBConfig} = require('../kyma-environment-broker');
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require('../compass');
const {oidcE2ETest, commerceMockTest} = require('./skr-test');
const {KCPWrapper, KCPConfig} = require('../kcp/client');
const t = require('../skr-svcat-migration-test/test-helpers');
const {keb, director} = require('./helpers');
const {initializeK8sClient} = require('../utils');
const {
  GardenerConfig,
  GardenerClient,
} = require('../gardener');
const {
  genRandom,
} = require('../utils');

const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe('Execute SKR test', function() {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 30; // 30m
  const deprovisioningTimeout = 1000 * 60 * 95; // 95m
  before('Provision SKR', async function() {
    try {
      this.options = gatherOptions();

      const keb = new KEBClient(KEBConfig.fromEnv());
      const gardener = new GardenerClient(GardenerConfig.fromEnv());
      const smAdminCreds = t.SMCreds.fromEnv();


      const suffix = genRandom(4);
      const runtimeName = `kyma-${suffix}`;
      const appName = `app-${suffix}`;

      const btpOperatorInstance = `btp-operator-${suffix}`;
      const btpOperatorBinding = `btp-operator-binding-${suffix}`;

      const btpOperatorCreds = await t.smInstanceBinding(smAdminCreds, btpOperatorInstance, btpOperatorBinding);

      const customParams = {
        oidc: this.options.oidc0,
      };

      console.log(`\nInstanceID ${this.options.instanceID}`,
          `Runtime ${runtimeName}`, `Application ${appName}`, `Suffix ${suffix}`);

      const skr = await provisionSKR(keb,
          kcp, gardener,
          this.options.instanceID,
          runtimeName,
          null,
          btpOperatorCreds,
          customParams,
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

  after('', async function() {
    // try {
    //   await deprovisionSKR(keb, kcp, this.options.instanceID, deprovisioningTimeout);
    // } catch (e) {
    //   throw new Error(`before hook failed: ${e.toString()}`);
    // } finally {
    //   const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
    //   console.log(`\nRuntime status after deprovisioning: ${runtimeStatus}`);
    //   await kcp.reconcileInformationLog(runtimeStatus);
    // }
    // await unregisterKymaFromCompass(director, this.options.scenarioName);
  });
});
