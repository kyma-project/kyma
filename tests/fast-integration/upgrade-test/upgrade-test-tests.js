const {
  checkInClusterEventDelivery,
  checkFunctionResponse,
  sendEventAndCheckResponse,
} = require("../test/fixtures/commerce-mock");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");
const {
  checkServiceInstanceExistence,
} = require("./fixtures/helm-broker");

describe("Upgrade test tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;
  const testNamespace = "test";

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it("in-cluster event should be delivered", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNamespace);
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("service instance provisioned by helm broker should be reachable", async function () {
    await checkServiceInstanceExistence(testNamespace);
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

});
