const {
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
} = require("../test/fixtures/commerce-mock");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");

describe("Upgrade test tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
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
