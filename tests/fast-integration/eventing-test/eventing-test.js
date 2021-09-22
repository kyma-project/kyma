const axios = require("axios");
const https = require("https");
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
} = require("../test/fixtures/commerce-mock");

describe("Eventing tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";
  const testStartTimestamp = new Date().toISOString();

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

  it("In-cluster event should be delivered (structured and binary mode)", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    await checkAppGatewayResponse();
  });

  it("order.created.v1 event from CommerceMock should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, true);
  });
});
