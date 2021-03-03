const {
  ensureGettingStartedTestFixture,
  verifyOrderPersisted,
  cleanGettingStartedTestFixture,
} = require("./fixtures/getting-started-guides");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
  shouldRunSuite,
} = require("../utils");

describe("Getting Started Guide Tests", function () {
  if(!shouldRunSuite("getting-started-guide")) {
    return;
  }

  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it("Getting started guide fixture should be ready", async function () {
    await ensureGettingStartedTestFixture().catch((err) => {
      console.dir(err); // first error is logged
      return ensureGettingStartedTestFixture();
    });
  });

  it("Order should be persisted and should survive pod restarts (redis storage)", async function () {
    await verifyOrderPersisted();
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  it("Namespace should be deleted", async function () {
    await cleanGettingStartedTestFixture(false);
  });
});
