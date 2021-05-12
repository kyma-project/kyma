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
} = require("../test/fixtures/commerce-mock");
const {
  debug,
  genRandom,
  initializeK8sClient,
} = require("../utils");

describe("Monitoring test", function() {

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  const testNS = "monitoring-test";

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);  
  
 
});