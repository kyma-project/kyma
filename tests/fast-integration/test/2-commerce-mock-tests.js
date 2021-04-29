const {
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
} = require("./fixtures/commerce-mock");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");

describe("CommerceMock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
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

});
