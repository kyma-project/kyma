const uuid = require("uuid");
const {
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
  updateSKR,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
  runtimes,
} = require("../../kyma-environment-broker");
const {
  DirectorConfig,
  DirectorClient,
  addScenarioInCompass,
  assignRuntimeToScenario,
  unregisterKymaFromCompass,
} = require("../../compass");
const { GardenerConfig, GardenerClient } = require("../../gardener");
const {
  ensureCommerceMockWithCompassTestFixture,
  checkFunctionResponse,
  sendEventAndCheckResponse,
  deleteMockTestFixture,
} = require("../../test/fixtures/commerce-mock");
const {
  debug,
  initializeK8sClient,
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
  getEnvOrThrow, sleep,
} = require("../../utils");

const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold,
} = require("../../audit-log");

const {
  KCPConfig,
  KCPWrapper,
} = require("../../kcp/client")
const {OIDCE2ETest, GatherOptions, WithRuntimeName, WithScenarioName, WithAppName, WithTestNS, CommerceMockTest} = require("../../skr-test");


describe("SKR nightly", function () {
  const keb = new KEBClient(KEBConfig.fromEnv());

  process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
  process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
  process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
  process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
  process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
  const kcp = new KCPWrapper(KCPConfig.fromEnv());

  const gardener = new GardenerClient(GardenerConfig.fromEnv());
  const director = new DirectorClient(DirectorConfig.fromEnv());

  const appName = `app-nightly`;
  const runtimeName = `kyma-nightly`;
  const scenarioName = `test-nightly`;
  const instanceID = uuid.v4();
  const oidc0 = {
    clientID: "abc-xyz",
    groupsClaim: "groups",
    issuerURL: "https://custom.ias.com",
    signingAlgs: ["RS256"],
    usernameClaim: "sub",
    usernamePrefix: "-",
  };

  const administrator0 = getEnvOrThrow("KEB_USER_ID");

  const oidc1 = {
    clientID: "foo-bar",
    groupsClaim: "groups1",
    issuerURL: "https://new.custom.ias.com",
    signingAlgs: ["RS256"],
    usernameClaim: "email",
    usernamePrefix: "acme-",
  };
  const administrators1 = ["admin1@acme.com", "admin2@acme.com"];

  debug(
    `RuntimeID ${instanceID}`,
    `Scenario ${scenarioName}`,
    `Runtime ${runtimeName}`,
    `Application ${appName}`
  );

  const testNS = "skr-nightly";
  const AWS_PLAN_ID = "361c511f-f939-4621-b228-d0fb79a1fe15";

  this.timeout(3600000 * 3); // 3h
  this.slow(5000);

  let skr;

  before(`Fetch previous nightly and deprovision if needed`, async function () {
    let runtime;
    console.log('Login to KCP.');
    await kcp.login()
    let query = {
      subaccount: keb.subaccountID,
    }
    try {
      console.log('Fetch last SKR.');
      let runtimes = await kcp.runtimes(query);
      if (runtimes.data) {
        runtime = runtimes.data[0];
      }
      if (runtime) {
        console.log('Deprovision last SKR.')
        await deprovisionSKR(keb, runtime.instanceID);
        await unregisterKymaFromCompass(director, scenarioName);
      } else {
        console.log("Deprovisioning not needed - no previous SKR found.");
      }

      console.log(`Provision SKR with runtime ID ${instanceID}`)
      const customParams = {
        oidc: oidc0,
      };

      skr = await provisionSKR(keb, gardener, instanceID, runtimeName, null, null, customParams);
      initializeK8sClient({ kubeconfig: skr.shoot.kubeconfig });
    }
     catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    }
  });
  let options = GatherOptions(
      WithRuntimeName('kyma-nightly'),
      WithScenarioName('test-nightly'),
      WithAppName('app-nightly'),
      WithTestNS('skr-nightly'));
  OIDCE2ETest(skr, instanceID, options);
  CommerceMockTest(skr, options);
});
