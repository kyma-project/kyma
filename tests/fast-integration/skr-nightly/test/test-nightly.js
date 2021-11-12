const {
   KCPConfig,
   KCPWrapper,
} = require('../../kcp/client');

const {initializeK8sClient} = require('../../utils');

const {
   GatherOptions, WithRuntimeName, WithScenarioName, WithAppName, WithTestNS, keb, gardener,
   OIDCE2ETest, CommerceMockTest, WithInstanceID,
} = require('../../skr-test');

// Mocha root hook
process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe(`SKR Nightly periodic test`, function () {
   this.timeout(60 * 60 * 1000 * 3); // 3h
   this.slow(5000);
   before('Fetch last nightly SKR', async function () {
      try {
         let runtime;
         await kcp.login();
         let query = {
            subaccount: keb.subaccountID,
         }
         console.log('Fetch last SKR.');
         let runtimes = await kcp.runtimes(query);
         if (runtimes.data) {
            runtime = runtimes.data[0];
         }
         this.shoot = await gardener.getShoot(runtime.shootName);
         this.options = GatherOptions(
             WithInstanceID(runtime.instanceID),
             WithRuntimeName('kyma-nightly'),
             WithScenarioName('test-nightly'),
             WithAppName('app-nightly'),
             WithTestNS('skr-nightly'));
         initializeK8sClient({ kubeconfig: this.shoot.kubeconfig });
      } catch (e) {
         throw new Error(`before hook failed: ${e.toString()}`);
      }
   });
   CommerceMockTest();
});
