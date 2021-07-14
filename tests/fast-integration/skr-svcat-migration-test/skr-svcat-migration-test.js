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
  getEnvOrThrow,
  initializeK8sClient,
  k8sCoreV1Api,
} = require("../utils");
const { 
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold
} = require("../audit-log");
const t = require("./test-helpers");

describe("SKR SVCAT migration test", function() {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const runtimeID = uuid.v4();
  
  const svcatPlatform = `svcat-${suffix}`
  const btpOperatorInstance = `btp-operator-${suffix}`
  const btpOperatorBinding = `btp-operator-binding-${suffix}`

  debug(`RuntimeID ${runtimeID}`, `Runtime ${runtimeName}`, `Application ${appName}`, `Suffix ${suffix}`);

  const testNS = "skr-test";
  const AWS_PLAN_ID = "361c511f-f939-4621-b228-d0fb79a1fe15";

  const clientID = getEnvOrThrow("BTP_OPERATOR_CLIENTID");
  const clientSecret = getEnvOrThrow("BTP_OPERATOR_CLIENTSECRET");
  const URL = getEnvOrThrow("BTP_OPERATOR_URL");

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);  

  let skr;
  it(`Should provision SKR`, async function() {
    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName);
  });
  it(`Should initialize K8s`, async function() {
    await initializeK8sClient({kubeconfig: skr.shoot.kubeconfig});
  });
  let creds;
  it(`Should instantiate SM Instance and Binding`, async function() {
    creds = await t.smInstanceBinding(URL, clientID, clientSecret, svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  });
  it(`Should install helm charts`, async function() {
    await t.installHelmCharts(k8sCoreV1Api, creds);
  });
  it(`Should deprovision SKR`, async function() {
    await deprovisionSKR(keb, runtimeID);
  });
  it(`Should cleanup SM instances and bindings`, async function() {
    await t.cleanupInstanceBinding(svcatPlatform, btpOperatorInstance, btpOperatorBinding);
  });
});
