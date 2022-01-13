const {gatherOptions} = require('./');
const {provisionSKR, deprovisionSKR} = require('../kyma-environment-broker');
const {keb, gardener, director} = require('./helpers');
const {initializeK8sClient} = require('../utils');
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require('../compass');
const {oidcE2ETest, commerceMockTest} = require('./skr-test');
const {KCPWrapper, KCPConfig} = require('../kcp/client');

const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe('Execute SKR test', function() {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 30; // 30m
  const deprovisioningTimeout = 1000 * 60 * 95; // 95m

  before('Provision SKR', async function() {
    try {
      this.options = gatherOptions();
      console.log(`Provision SKR with instance ID ${this.options.instanceID}`);
      const customParams = {
        oidc: this.options.oidc0,
      };
      const skr = await provisionSKR(keb, kcp, gardener,
          this.options.instanceID,
          this.options.runtimeName,
          null,
          null,
          customParams,
          provisioningTimeout);

      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);

      this.shoot = skr.shoot;
      await addScenarioInCompass(director, this.options.scenarioName);
      await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
      initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    }
  });

  oidcE2ETest();
  commerceMockTest();

  after('Deprovision SKR', async function() {
    await deprovisionSKR(keb, kcp, this.options.instanceID, deprovisioningTimeout);
    await unregisterKymaFromCompass(director, this.options.scenarioName);
  });
});
