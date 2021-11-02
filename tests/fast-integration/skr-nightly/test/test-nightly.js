const {
   KCPConfig,
   KCPWrapper,
} = require('../../kcp/client');

const {
   KEBConfig,
   KEBClient,
} = require('../../kyma-environment-broker');

const {
   GardenerClient,
   GardenerConfig,
} = require('../../gardener');

const {initializeK8sClient} = require('../../utils');

const {
   commerceMockTests,
   gettingStartedGuides,
} = require('../../test');

describe(`SKR Nightly periodic test`, function () {
   const kebconfig = KEBConfig.fromEnv();
   const gardenerConfig = GardenerConfig.fromEnv();

   const gardener = new GardenerClient(gardenerConfig);
   const keb = new KEBClient(kebconfig);

   process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
   process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
   process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;

   const config = KCPConfig.fromEnv();
   const kcp = new KCPWrapper(config);

   let runtime;

   describe(`Prepare kube client`, function () {
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
         const shoot = await gardener.getShoot(runtime.shootName);
         console.log(shoot.name, runtime.shootName);
         initializeK8sClient({ kubeconfig: shoot.kubeconfig });
      });
   });
   commerceMockTests();
   gettingStartedGuides();
});
