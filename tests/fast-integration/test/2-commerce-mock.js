const axios = require("axios");
const https = require("https");
const { expect } = require("chai");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  ensureCommerceMockLocalTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
  cleanMockTestFixture,
  checkInClusterEventDelivery,
} = require("./fixtures/commerce-mock");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");

describe("CommerceMock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const withCentralApplicationGateway = process.env.WITH_CENTRAL_APPLICATION_GATEWAY || false;
  const testNamespace = "test";
  let initialRestarts = null;

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace, withCentralApplicationGateway).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace, withCentralApplicationGateway);
    });
  });

  it("in-cluster event should be delivered", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    await checkAppGatewayResponse();
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, true);
  });
});
