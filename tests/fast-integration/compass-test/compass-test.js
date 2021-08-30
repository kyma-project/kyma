import { DirectorConfig, DirectorClient, registerKymaInCompass, unregisterKymaFromCompass } from "../compass/";
import { genRandom } from "../utils";
import { ensureCommerceMockWithCompassTestFixture, cleanMockTestFixture, checkFunctionResponse, sendEventAndCheckResponse } from "../test/fixtures/commerce-mock";

describe("Kyma with Compass test", async function() {
  const director = new DirectorClient(DirectorConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;

  const testNS = "compass-test";

  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  it("Register Kyma instance in Compass", async function() {
    await registerKymaInCompass(director, runtimeName, scenarioName);
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockWithCompassTestFixture(director, appName, scenarioName,  "mocks", testNS);
  });

  it("function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNS);
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