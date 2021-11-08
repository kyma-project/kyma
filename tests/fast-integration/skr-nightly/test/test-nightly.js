const {
   KCPConfig,
   KCPWrapper,
} = require('../../kcp/client');

const {initializeK8sClient} = require('../../utils');

const {
   GatherOptions, WithRuntimeName, WithScenarioName, WithAppName, WithTestNS, keb, gardener,
   OIDCE2ETest, CommerceMockTest,
} = require('../../skr-test');


describe(`SKR Nightly periodic test`, function () {
   process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
   process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
   process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
   process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';

   const config = KCPConfig.fromEnv();
   const kcp = new KCPWrapper(config);

   let instanceID;
   let shoot;
   before('Fetch last nightly SKR', async function () {
      let runtime;
      await kcp.login()
      let query = {
         subaccount: keb.subaccountID,
      }
      let runtimes = await kcp.runtimes(query);
      if (runtimes.data) {
         runtime = runtimes.data[0];
         instanceID = runtime.instanceID;
      }
      console.log(runtime);
      shoot = await gardener.getShoot(runtime.shootName);
      initializeK8sClient({ kubeconfig: shoot.kubeconfig });
   });
   describe('Execute tests', function () {
      let options = GatherOptions(
          WithRuntimeName('kyma-nightly'),
          WithScenarioName('test-nightly'),
          WithAppName('app-nightly'),
          WithTestNS('skr-nightly'));
      let skr = {
         operation: "",
         shoot
      }
      OIDCE2ETest(skr, instanceID, options);
      CommerceMockTest(skr, options);
   });
});
