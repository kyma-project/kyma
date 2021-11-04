const {
   KCPConfig,
   KCPWrapper,
} = require('../../kcp/client');

const {initializeK8sClient} = require('../../utils');

const {
   skrTest, GatherOptions, WithRuntimeID, WithRuntimeName, WithScenarioName, WithAppName, WithTestNS, keb, gardener,
} = require('../../skr-test');

describe(`SKR Nightly periodic test`, function () {
   process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
   process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
   process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;

   const config = KCPConfig.fromEnv();
   const kcp = new KCPWrapper(config);

   let options;
   let runtime;
   let shoot;

   describe(`Prepare step`, function () {
      it(`Fetch last runtime`, async function() {
         await kcp.login()
         let query = {
            subaccount: keb.subaccountID,
         }
         let runtimes = await kcp.runtimes(query);
         if (runtimes.data) {
            runtime = runtimes.data[0];
         }
      });
      it (`Initialize k8s client from nightly runtime`, async function () {
         shoot = await gardener.getShoot(runtime.shootName);
         initializeK8sClient({ kubeconfig: shoot.kubeconfig });
      });
      it('Initialize test options for nightly', async function () {
         options = GatherOptions(
             WithRuntimeID(runtime.instanceID),
             WithRuntimeName('kyma-nightly'),
             WithScenarioName('test-nightly'),
             WithAppName('app-nightly'),
             WithTestNS('skr-nightly'));
         console.log(options)
      })
   });
   let skr = {
      operation: "",
      shoot
   }
   // skrTest(skr, options);
});
