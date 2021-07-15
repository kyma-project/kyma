const uuid = require("uuid");
const { 
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
} = require("../kyma-environment-broker");
const {
  DirectorConfig,
  DirectorClient,
  addScenarioInCompass,
  assignRuntimeToScenario,
  unregisterKymaFromCompass,
} = require("../compass");
const {
  GardenerConfig,
  GardenerClient
} = require("../gardener");
const {
  ensureCommerceMockWithCompassTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
  deleteMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
  debug,
  genRandom,
  initializeK8sClient,
} = require("../utils");
const { 
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold
} = require("../audit-log");
const t = require("./install-helm-charts");

describe("SKR SVCAT migration test", function() {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  const testNS = "skr-test";
  const AWS_PLAN_ID = "361c511f-f939-4621-b228-d0fb79a1fe15";

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);  

  let skr;
  
  it(`Provision SKR with ID ${runtimeID}`, async function() {
    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName);
    initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });

  it(`Create SM instance and binding for BTP Operator`, async function() {
    creds = await t.smInstanceBinding();
  });

  it(`Install helm charts`, async function() {
    await t.installHelmCharts(creds);
  });
   
  it("Deprovision SKR", async function() {
    await deprovisionSKR(keb, runtimeID);
  });
});
