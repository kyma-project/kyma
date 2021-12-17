const {
  provisionSKR,
  deprovisionSKR,
} = require('../../kyma-environment-broker');
const {
  unregisterKymaFromCompass,
  addScenarioInCompass,
  assignRuntimeToScenario,
} = require('../../compass');
const {
  initializeK8sClient,
} = require('../../utils');

const {
  KCPConfig,
  KCPWrapper,
} = require('../../kcp/client');
const {
  oidcE2ETest,
  gatherOptions,
  withRuntimeName,
  withScenarioName,
  commerceMockTest,
  gardener,
  keb,
  director,
} = require('../../skr-test');

// Mocha root hook
process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
const kcp = new KCPWrapper(KCPConfig.fromEnv());


describe('SKR nightly', function() {
  this.timeout(3600000 * 3); // 3h
  this.slow(5000);
  before(`Fetch last SKR and deprovision if needed`, async function() {
    try {
      let runtime;
      this.options = gatherOptions(
          withRuntimeName('kyma-nightly'),
          withScenarioName('test-nightly'));

      console.log('Login to KCP.');
      await kcp.login();
      const query = {
        subaccount: keb.subaccountID,
      };
      console.log('Fetch last SKR.');
      const runtimes = await kcp.runtimes(query);
      if (runtimes.data) {
        runtime = runtimes.data[0];
      }
      if (runtime) {
        console.log('Deprovision last SKR.');
        await deprovisionSKR(keb, runtime.instanceID);
        await unregisterKymaFromCompass(director, this.options.scenarioName);
      } else {
        console.log('Deprovisioning not needed - no previous SKR found.');
      }

      console.log(`Provision SKR with runtime ID ${this.options.instanceID}`);
      const customParams = {
        oidc: this.options.oidc0,
      };

      const skr = await provisionSKR(keb,
          gardener,
          this.options.instanceID,
          this.options.runtimeName,
          null,
          null,
          customParams);
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
});
