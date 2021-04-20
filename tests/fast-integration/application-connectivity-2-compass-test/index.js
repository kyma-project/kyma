const { 
  DirectorConfig, 
  DirectorClient,
  registerKymaInCompass,
  unregisterKymaFromCompass,
} = require("../compass/");

const {
  genRandom, debug
} = require("../utils");

const {
  ensureCommerceMockWithCompassTestFixture,
  cleanMockTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse
} = require("../test/fixtures/commerce-mock");

const installer = require("../installer");

describe("Kyma with Compass test", async function() {
  const director = new DirectorClient(DirectorConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  
  debug(`Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  const testNS = "compass-test";
  const mockNamespace = "mocks";
  const skipComponents = ["dex","tracing","monitoring","console","kiali","logging"];
  const newEventing = true;
  const withCompass = true;
  const withCentralApplicationGateway = true;

  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  it("Install Kyma", async function() {
    await installer.installKyma({newEventing, withCompass, withCentralApplicationGateway, skipComponents});
  });

  it("Register Kyma instance in Compass", async function() {
    await registerKymaInCompass(director, runtimeName, scenarioName);
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockWithCompassTestFixture(director, appName, scenarioName,  "mocks", testNS, withCentralApplicationGateway);
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    await checkAppGatewayResponse();
  });
    
  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Unregister Kyma resources from Compass", async function() {
    await unregisterKymaFromCompass(director, scenarioName);
  });

  it("Test fixtures should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNS, true)
  });
});