const {
  KCPConfig,
  KCPWrapper,
} = require('../../kcp/client');

const {initializeK8sClient} = require('../../utils');

const {
  gatherOptions,
  withRuntimeName,
  withScenarioName,
  keb,
  gardener,
  withInstanceID,
  commerceMockTest,
} = require('../../skr-test');

// Mocha root hook
process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe('SKR Nightly periodic test', function() {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  let options;
  let shoot;

  before('Fetch last nightly SKR', async function() {
    try {
      let runtime;
      await kcp.login();
      const query = {
        subaccount: keb.subaccountID,
      };
      console.log('Fetch last SKR.');
      const runtimes = await kcp.runtimes(query);
      if (runtimes.data) {
        runtime = runtimes.data[0];
      }
      shoot = await gardener.getShoot(runtime.shootName);
      options = gatherOptions(
          withInstanceID(runtime.instanceID),
          withRuntimeName('kyma-nightly'),
          withScenarioName('test-nightly'));
      initializeK8sClient({kubeconfig: shoot.kubeconfig});
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    }
  });
  commerceMockTest(options);
});
