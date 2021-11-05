const {
   KCPConfig,
   KCPWrapper,
} = require('../../kcp/client');

const {initializeK8sClient} = require('../../utils');

const {
   GatherOptions, WithRuntimeID, WithRuntimeName, WithScenarioName, WithAppName, WithTestNS, keb, gardener,
   OIDCE2ETest, CommerceMockTest,
} = require('../../skr-test');


describe(`SKR Nightly periodic test`, function () {
   process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
   process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
   process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;

   const config = KCPConfig.fromEnv();
   const kcp = new KCPWrapper(config);

   let runtime;
   let shoot;
   before('Provision SKR', async function () {
      await kcp.login()
      let query = {
         subaccount: keb.subaccountID,
      }
      let runtimes = await kcp.runtimes(query);
      if (runtimes.data) {
         runtime = runtimes.data[0];
      }
      console.log(runtime);
      shoot = await gardener.getShoot(runtime.shootName);
      initializeK8sClient({ kubeconfig: shoot.kubeconfig });
   });
   describe('Execute tests', function () {
      let options = GatherOptions(
          WithRuntimeID(runtime.instanceID),
          WithRuntimeName('kyma-nightly'),
          WithScenarioName('test-nightly'),
          WithAppName('app-nightly'),
          WithTestNS('skr-nightly'));
      let skr = {
         operation: "",
         shoot
      }
      console.log(options);
      OIDCE2ETest(skr, options);
      CommerceMockTest(skr, options);
   });
});
