const uuid = require("uuid");
const fs = require('fs');
const execa = require("execa");
const { 
  DirectorConfig, 
  DirectorClient,
  addScenarioInCompass,
  assignRuntimeToScenario,
  unregisterKymaFromCompass,
} = require("../compass");
const {
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
} = require("../kyma-environment-broker");
const { GardenerConfig, GardenerClient } = require("../gardener");
const {
  debug,
  genRandom,
  initializeK8sClient,
  printRestartReport,
  getContainerRestartsForAllNamespaces,
  getEnvOrThrow,
  switchDebug,
} = require("../utils");
const {
  ensureCommerceMockWithCompassTestFixture,
  checkInClusterEventDelivery,
  checkFunctionResponse,
  sendEventAndCheckResponse,
} = require("../test/fixtures/commerce-mock");
const {
  checkServiceInstanceExistence,
  ensureHelmBrokerTestFixture,
} = require("../upgrade-test/fixtures/helm-broker");
const {
  cleanMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
  KCPConfig,
  KCPWrapper,
} = require("../kcp/client")
const {
  saveKubeconfig,
} = require("../skr-svcat-migration-test/test-helpers");

describe("SKR-Upgrade-test", function () {
  switchDebug(on = true)
  let keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const instanceID = uuid.v4();
  const subAccountID = uuid.v4();

  keb.subaccountID = subAccountID;

  debug(
    `PlanID ${getEnvOrThrow("KEB_PLAN_ID")}`,
    `SubAccountID ${subAccountID}`,
    `InstanceID ${instanceID}`,
    `Scenario ${scenarioName}`,
    `Runtime ${runtimeName}`,
    `Application ${appName}`
  );

  // debug(
  //   `KEB_HOST: ${getEnvOrThrow("KEB_HOST")}`,
  //   `KEB_CLIENT_ID: ${getEnvOrThrow("KEB_CLIENT_ID")}`,
  //   `KEB_CLIENT_SECRET: ${getEnvOrThrow("KEB_CLIENT_SECRET")}`,
  //   `KEB_GLOBALACCOUNT_ID: ${getEnvOrThrow("KEB_GLOBALACCOUNT_ID")}`,
  //   `KEB_SUBACCOUNT_ID: ${getEnvOrThrow("KEB_SUBACCOUNT_ID")}`,
  //   `KEB_USER_ID: ${getEnvOrThrow("KEB_USER_ID")}`,
  //   `KEB_PLAN_ID: ${getEnvOrThrow("KEB_PLAN_ID")}`      
  // );
  
  // debug(
  //   `COMPASS_HOST: ${getEnvOrThrow("COMPASS_HOST")}`,
  //   `COMPASS_CLIENT_ID: ${getEnvOrThrow("COMPASS_CLIENT_ID")}`,
  //   `COMPASS_CLIENT_SECRET: ${getEnvOrThrow("COMPASS_CLIENT_SECRET")}`,
  //   `COMPASS_TENANT: ${getEnvOrThrow("COMPASS_TENANT")}`,
  // )
  
  // debug(
  //   `KCP_TECH_USER_LOGIN: ${KCP_TECH_USER_LOGIN}\n`,
  //   `KCP_TECH_USER_PASSWORD: ${KCP_TECH_USER_PASSWORD}\n`,
  //   `KCP_OIDC_CLIENT_ID: ${KCP_OIDC_CLIENT_ID}\n`,
  //   `KCP_OIDC_CLIENT_SECRET: ${KCP_OIDC_CLIENT_SECRET}\n`,
  //   `KCP_KEB_API_URL: ${KCP_KEB_API_URL}\n`,
  //   `KCP_OIDC_ISSUER_URL: ${KCP_OIDC_ISSUER_URL}\n`
  // )

  // Credentials for KCP ODIC Login
  
  // process.env.KCP_TECH_USER_LOGIN    = 
  // process.env.KCP_TECH_USER_PASSWORD = 
  process.env.KCP_OIDC_ISSUER_URL    = "https://kymatest.accounts400.ondemand.com";
  // process.env.KCP_OIDC_CLIENT_ID     = 
  // process.env.KCP_OIDC_CLIENT_SECRET = 
  process.env.KCP_KEB_API_URL        = "https://kyma-env-broker.cp.dev.kyma.cloud.sap";
  process.env.KCP_GARDENER_NAMESPACE = "garden-kyma-dev";
  process.env.KCP_MOTHERSHIP_API_URL = "https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1"
  process.env.KCP_KUBECONFIG_API_URL = "https://kubeconfig-service.cp.dev.kyma.cloud.sap"
  
  const kcp = new KCPWrapper(KCPConfig.fromEnv());

  const kymaUpgradeVersion = getEnvOrThrow("KYMA_UPGRADE_VERSION")

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  let skr;

  // SKR Provisioning

  it(`Provision SKR with ID ${instanceID}`, async function () {
    skr = await provisionSKR(keb, gardener, instanceID, runtimeName, null, null, null);
  });

  it(`Should save kubeconfig for the SKR to ~/.kube/config`, async function() {
    saveKubeconfig(skr.shoot.kubeconfig);
  });

  it(`Should initialize K8s client`, async function() {
    await initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });

  // Upgrade Test Praparation
  const director = new DirectorClient(DirectorConfig.fromEnv());
  const withCentralAppConnectivity = (process.env.WITH_CENTRAL_APP_CONNECTIVITY === "true");

  const testNS = "test";

  it("Assign SKR to scenario", async function () {
    await addScenarioInCompass(director, scenarioName);
    await assignRuntimeToScenario(director, skr.shoot.compassID, scenarioName);
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockWithCompassTestFixture(director, appName, scenarioName,  "mocks", testNS, withCentralAppConnectivity);
  });

  it("Helm Broker test fixture should be ready", async function () {
    await ensureHelmBrokerTestFixture(testNS).catch((err) => {
      console.dir(err); // first error is logged
      return ensureHelmBrokerTestFixture(testNS);
    });
  });

  // Perform Tests before Upgrade

  it("Listing all pods in cluster", async function () {
    await getContainerRestartsForAllNamespaces();
  });

  let initialRestarts

  it("in-cluster event should be delivered", async function () {
    initialRestarts = await checkInClusterEventDelivery(testNS);
  });

  it("function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNS);
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("service instance provisioned by helm broker should be reachable", async function () {
    await checkServiceInstanceExistence(testNS);
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  // Perform Upgrade
  
  it(`Perform kcp login`, async function () {
    
    let version = await kcp.version([])
    debug(version)
    
    await kcp.login()
    // debug(loginOutput)
  });


  it(`Perform Upgrade`, async function () {
    let kcpUpgradeStatus = await kcp.upgradeKyma(subAccountID, kymaUpgradeVersion)
    debug("Upgrade Done!")
  });

  // Perform Tests after Upgrade

  it("Listing all pods in cluster", async function () {
    await getContainerRestartsForAllNamespaces();
  });

  initialRestarts = undefined

  it("in-cluster event should be delivered", async function () {
    initialRestarts = await checkInClusterEventDelivery(testNS);
  });

  it("function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNS);
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("service instance provisioned by helm broker should be reachable", async function () {
    await checkServiceInstanceExistence(testNS);
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });


  // Cleanup
  const skip_cleanup = getEnvOrThrow("SKIP_CLEANUP")

  if (skip_cleanup === "FALSE") {
    it("Unregister Kyma resources from Compass", async function() {
      await unregisterKymaFromCompass(director, scenarioName);
    });

    it("Test fixtures should be deleted", async function () {
      await cleanMockTestFixture("mocks", testNS, true)
    });

    it("Unregister SKR resources from Compass", async function () {
      await unregisterKymaFromCompass(director, scenarioName);
    });

    it("Deprovision SKR", async function () {
      await deprovisionSKR(keb, instanceID);
    });
  }
});
