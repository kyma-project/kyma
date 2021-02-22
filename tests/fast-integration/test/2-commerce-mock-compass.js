const {
    ensureCommerceMockTestFixture,
    checkAppGatewayResponse,
    sendEventAndCheckResponse,
    cleanMockTestFixture,
    connectMockCompass,
    registerKymaInCompass,
    unregisterKymaFromCompass,
} = require("./fixtures/commerce-mock");

const {
  genRandom,
  shouldRunSuite,
} = require("../utils");

const {
    DirectorClient
} = require("../compass");

describe("CommerceMock with Compass tests", function () {
  if(!shouldRunSuite("commerce-mock-compass")) {
    return;
  }

    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const testNamespace = "compass-test";
  
    const compassHost = process.env["COMPASS_HOST"] || "";
    const clientID = process.env["COMPASS_CLIENT_ID"] || "";
    const clientSecret = process.env["COMPASS_CLIENT_SECRET"] || "";
    const tenantID = process.env["COMPASS_TENANT"] || "";

    const director = new DirectorClient(compassHost, clientID, clientSecret, tenantID);

    const suffix = genRandom(4);
    const appName = `commerce-${suffix}`;
    const runtimeName = `kyma-${suffix}`;
    const scenarioName = `integration-${suffix}`;

    console.log(runtimeName, appName, scenarioName);
    
    it("Register Kyma in Compass", async function() {
      await registerKymaInCompass(director, runtimeName, scenarioName);
    })
  
    it("CommerceMock test fixture should be ready", async function () {
      const connectFn = connectMockCompass(director, appName, scenarioName, testNamespace);
      await ensureCommerceMockTestFixture("mocks", testNamespace, connectFn).catch((err) => {
        console.dir(err); // first error is logged
        return ensureCommerceMockTestFixture("mocks", testNamespace, connectFn);
      });
    });
  
    // it("function should reach Commerce mock API through app gateway", async function () {
    //   await checkAppGatewayResponse();
    // });
  
    // it("order.created.v1 event should trigger the lastorder function", async function () {
    //   await sendEventAndCheckResponse();
    // });

    it("Unregister Kyma from Compass", async function() {
      await unregisterKymaFromCompass(director, scenarioName);
    });
  
    it("Test namespaces should be deleted", async function () {
      await cleanMockTestFixture("mocks", testNamespace, false);
    });
});
