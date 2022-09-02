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
  gatherOptions,
  withRuntimeName,
  withScenarioName,
  gardener,
  keb,
  director,
  oidcE2ETest,
} = require('../../skr-test');

// Mocha root hook
process.env.KCP_KEB_API_URL = 'https://kyma-env-broker.' + keb.host;
process.env.KCP_GARDENER_NAMESPACE = 'garden-kyma-dev';
process.env.KCP_OIDC_ISSUER_URL = 'https://kymatest.accounts400.ondemand.com';
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
const kcp = new KCPWrapper(KCPConfig.fromEnv());


describe('SKR nightly', function() {
  this.timeout(3600000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 60; // 1h
  const deprovisioningTimeout = 1000 * 60 * 30; // 30m
  let shoot;
  const getShootInfoFunc = function() {
    return shoot;
  };
  const options = gatherOptions(
      withRuntimeName('kyma-nightly'),
      withScenarioName('test-nightly'));
  const getShootOptionsFunc = function() {
    return options;
  };

  before('Fetch last SKR and deprovision if needed', async function() {
    try {
      let runtime;

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
        await deprovisionSKR(keb, kcp, runtime.instanceID, deprovisioningTimeout);
        await unregisterKymaFromCompass(director, options.scenarioName);
      } else {
        console.log('Deprovisioning not needed - no previous SKR found.');
      }

      console.log(`Provision SKR with runtime ID ${options.instanceID}`);
      const customParams = {
        oidc: options.oidc0,
      };

      const skr = await provisionSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          options.runtimeName,
          null,
          null,
          customParams,
          provisioningTimeout);

      const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
      console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);

      shoot = skr.shoot;
      await addScenarioInCompass(director, options.scenarioName);
      await assignRuntimeToScenario(director, shoot.compassID, options.scenarioName);
      initializeK8sClient({kubeconfig: shoot.kubeconfig});
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    }
  });

  oidcE2ETest(getShootOptionsFunc, getShootInfoFunc);
});
